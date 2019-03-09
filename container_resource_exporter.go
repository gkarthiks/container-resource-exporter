package main

import (
	_ "expvar"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsTypes "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	kubeconfig        *string
	pods              *corev1.PodList
	clientset         *v1.Clientset
	metricClientSet   *metrics.Clientset
	watchNamespace    string
	err               error
	avail             bool
	podCounts         = make(map[string]int)
	podCountsMapMutex = sync.RWMutex{}
)

// Exporter collects metrics and exports them using
// the prometheus metrics package.
type Exporter struct {
	cpuRequest    *prometheus.Desc
	memoryRequest *prometheus.Desc
	cpuLimit      *prometheus.Desc
	memoryLimit   *prometheus.Desc
	cpuUsage      *prometheus.Desc
	memoryUsage   *prometheus.Desc
	totalPods     *prometheus.Desc
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	watchNamespace, avail = os.LookupEnv("WATCH_NAMESPACE")
	if avail {
		logrus.Infof("Chosen namespace to scrape: %s", watchNamespace)
	} else {
		logrus.Info("No watch namespace provided, defaulting to cluster level")
		watchNamespace = ""
	}

	// uncomment below, if running outside cluster
	// if home := homeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }

	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// if running inside cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// k8s core api client
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// k8s metrics api client
	metricClientSet, err = metrics.NewForConfig(config)
	if err != nil {
		panic(err)
	}

}

func main() {
	resExporter := NewExporter()
	prometheus.MustRegister(resExporter)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>CRE</title></head>
             <body>
             <h1>Container Resource Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
	http.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
		logrus.Info("Running healthz check")
		io.WriteString(rw, "Running good")
	})
	logrus.Info("Serving on port :9000")
	logrus.Fatal(http.ListenAndServe(":9000", nil))
}

// NewExporter initializes every descriptor and returns a pointer to the collector
func NewExporter() *Exporter {
	return &Exporter{
		cpuRequest: prometheus.NewDesc("cpu_request",
			"Requested CPU by deployment",
			[]string{"pod_name", "container_name", "namespace", "status"}, nil,
		),
		memoryRequest: prometheus.NewDesc("memory_request",
			"Requested Memory by deployment",
			[]string{"pod_name", "container_name", "namespace", "status"}, nil,
		),
		cpuLimit: prometheus.NewDesc("cpu_limit",
			"CPU Limit by deployment",
			[]string{"pod_name", "container_name", "namespace", "status"}, nil,
		),

		memoryLimit: prometheus.NewDesc("memory_limit",
			"Memory Limit by deployment",
			[]string{"pod_name", "container_name", "namespace", "status"}, nil,
		),
		cpuUsage: prometheus.NewDesc("current_cpu_usage",
			"Current CPU Usage as reported by Metrics API",
			[]string{"pod_name", "container_name", "namespace"}, nil,
		),
		memoryUsage: prometheus.NewDesc("current_memory_usage",
			"Current CPU Usage as reported by Metrics API",
			[]string{"pod_name", "container_name", "namespace"}, nil,
		),
		totalPods: prometheus.NewDesc("total_pod",
			"Total pod count in given space",
			[]string{"namespace"}, nil,
		),
	}
}

// Describe writes all descriptors to the prometheus desc channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cpuRequest
	ch <- e.cpuLimit
	ch <- e.cpuUsage

	ch <- e.memoryRequest
	ch <- e.memoryLimit
	ch <- e.memoryUsage

	ch <- e.totalPods
}

//Collect implements required collect function for all promehteus collectors
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	logrus.Infof("Scraping core and metrics api for metrics")
	podCounts = make(map[string]int)
	// var totalCPU, totalMemory float64
	var wg = sync.WaitGroup{}

	// Polling core API
	pods, err = clientset.CoreV1().Pods(watchNamespace).List(metav1.ListOptions{})
	if err != nil {
		logrus.Error(err.Error())
	}

	podCountNamespace := make(chan string)

	getPodDefinedresource := func(pod corev1.Pod) {
		defer wg.Done()
		for _, container := range pod.Spec.Containers {
			cpuRequestFloat, _ := strconv.ParseFloat(container.Resources.Requests.Cpu().AsDec().String(), 64)
			ch <- prometheus.MustNewConstMetric(e.cpuRequest, prometheus.GaugeValue, cpuRequestFloat, pod.Name, container.Name, pod.Namespace, fmt.Sprintf("%v ", pod.Status.Phase))

			memoryRequestFloat, _ := strconv.ParseFloat(container.Resources.Requests.Memory().AsDec().String(), 64)
			ch <- prometheus.MustNewConstMetric(e.memoryRequest, prometheus.GaugeValue, memoryRequestFloat, pod.Name, container.Name, pod.Namespace, fmt.Sprintf("%v ", pod.Status.Phase))

			cpuLimitFloat, _ := strconv.ParseFloat(container.Resources.Limits.Cpu().AsDec().String(), 64)
			ch <- prometheus.MustNewConstMetric(e.cpuLimit, prometheus.GaugeValue, cpuLimitFloat, pod.Name, container.Name, pod.Namespace, fmt.Sprintf("%v ", pod.Status.Phase))

			memoryLimitFloat, _ := strconv.ParseFloat(container.Resources.Limits.Memory().AsDec().String(), 64)
			ch <- prometheus.MustNewConstMetric(e.memoryLimit, prometheus.GaugeValue, memoryLimitFloat, pod.Name, container.Name, pod.Namespace, fmt.Sprintf("%v ", pod.Status.Phase))
		}
		podCountNamespace <- pod.Namespace
	}

	for _, pod := range pods.Items {
		wg.Add(1)
		go getPodDefinedresource(pod)
	}

	go setPodCount(podCountNamespace)

	podMetrics, err := metricClientSet.MetricsV1beta1().PodMetricses(watchNamespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	var podMetric metricsTypes.PodMetrics

	getPodUsageMetrics := func(pod metricsTypes.PodMetrics) {
		defer wg.Done()
		for _, container := range pod.Containers {
			cpuQuantityDec := container.Usage.Cpu().AsDec().String()
			cpuUsageFloat, _ := strconv.ParseFloat(cpuQuantityDec, 64)
			ch <- prometheus.MustNewConstMetric(e.cpuUsage, prometheus.GaugeValue, cpuUsageFloat, pod.Name, container.Name, pod.Namespace)

			memoryQuantityDec := container.Usage.Memory().AsDec().String()
			memoryUsageFloat, _ := strconv.ParseFloat(memoryQuantityDec, 64)
			ch <- prometheus.MustNewConstMetric(e.memoryUsage, prometheus.GaugeValue, memoryUsageFloat, pod.Name, container.Name, pod.Namespace)
		}
	}

	for _, podMetric = range podMetrics.Items {
		wg.Add(1)
		go getPodUsageMetrics(podMetric)
	}

	getPosNamespaceCount := func(namespace string, count int) {
		defer wg.Done()
		ch <- prometheus.MustNewConstMetric(e.totalPods, prometheus.CounterValue, float64(count), namespace)
	}

	for namespace, count := range podCounts {
		wg.Add(1)
		go getPosNamespaceCount(namespace, count)
	}
	wg.Wait()
	close(podCountNamespace)
}

func setPodCount(podCountsNamespace chan string) {
	for namespace := range podCountsNamespace {
		podCountsMapMutex.Lock()
		if count, ok := podCounts[namespace]; ok {
			count = count + 1
			podCounts[namespace] = count
		} else {
			podCounts[namespace] = 1
		}
		podCountsMapMutex.Unlock()
	}

}

// Uncomment if running outside the cluster {fetches the local kubeconfig}
// func homeDir() string {
// 	if h := os.Getenv("HOME"); h != "" {
// 		return h
// 	}
// 	return os.Getenv("USERPROFILE") // windows
// }
