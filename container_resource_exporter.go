package main

import (
	"context"
	_ "expvar"
	"fmt"
	discovery "github.com/gkarthiks/k8s-discovery"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsTypes "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	k8s               *discovery.K8s
	pods              *corev1.PodList
	podMetrics        *metricsTypes.PodMetricsList
	watchNamespace    string
	nsSlice           []string
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
	k8s, _ = discovery.NewK8s()
	watchNamespace, avail = os.LookupEnv("WATCH_NAMESPACE")
	if avail {
		logrus.Infof("Chosen namespace to scrape: %s", watchNamespace)
		if strings.Contains(watchNamespace, ",") {
			splitNamespace := strings.Split(watchNamespace, ",")
			for _, indNs := range splitNamespace {
				nsSlice = append(nsSlice, strings.TrimSpace(indNs))
			}
		}
	} else {
		logrus.Info("No watch namespace provided, defaulting to cluster level")
		watchNamespace = ""
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
		_, _ = io.WriteString(rw, "Running good")
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
	if len(nsSlice) > 0 {
		var podSlices = &corev1.PodList{}
		for _, namespace := range nsSlice {
			logrus.Infof("Currently scrapping the %s namespace", namespace)
			pods, err = k8s.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
			podSlices.Items = append(podSlices.Items, pods.Items...)
		}
		pods = podSlices
	} else {
		pods, err = k8s.Clientset.CoreV1().Pods(watchNamespace).List(context.Background(), metav1.ListOptions{})
	}
	if err != nil {
		logrus.Error(err.Error())
	} else if pods != nil && len(pods.Items) > 0 {
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

		if len(nsSlice) > 0 {
			var podMetricsSlices = &metricsTypes.PodMetricsList{}
			for _, namespace := range nsSlice {
				logrus.Infof("Currently scrapping the %s namespace for metrics", namespace)
				podMetrics, err = k8s.MetricsClientSet.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{})
				podMetricsSlices.Items = append(podMetricsSlices.Items, podMetrics.Items...)
			}
			podMetrics = podMetricsSlices
		} else {
			logrus.Println("Inside else using watchnamespace")
			podMetrics, err = k8s.MetricsClientSet.MetricsV1beta1().PodMetricses(watchNamespace).List(context.Background(), metav1.ListOptions{})
		}

		if err != nil {
			noRBACError := regexp.MustCompile(`.*.metrics.* is forbidden:.* cannot list.* no RBAC policy matched`)
			if noRBACError.MatchString(err.Error()) {
				logrus.Fatalf("The service account running this pod doesn't have a matching RBAC to fetch the Metrics, errored out: %v", err.Error())
			}
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
	} else {
		logrus.Infoln("No pod was listed to fetch the metrics")
	}
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
