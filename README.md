# container-resource-exporter
![Build Type](https://img.shields.io/docker/cloud/automated/gkarthics/container-resource-exporter.svg)
![Build Status](https://img.shields.io/docker/cloud/build/gkarthics/container-resource-exporter.svg)
![License](https://img.shields.io/github/license/gkarthiks/container-resource-exporter.svg)
![Release](https://img.shields.io/github/tag-date/gkarthiks/container-resource-exporter.svg?color=Orange&label=Latest%20Release)

Container Resource Exporter (CRE) is a metrics expoerter which will provide the *container* resource `request/limit/usage` metrics data on realtime in the [Prometheus](https://prometheus.io/) format. This can be utilized to trigger a pro-active alert from the the [Prometheus Alert Manager](https://prometheus.io/docs/alerting/alertmanager).


## Cluster Level:
By default the *CRE* will watch for the entire cluster, scrapes the resources for each and every container and exports them along with the total count of pods in each namesapces.

To run in the *cluster mode*, the *CRE* will require a *service account* which has cluster level read access and `list` actions on the resource `pods` in `core v1` apiGroup and `metrics` apiGroup.


## Contained Namespace:
To run the *CRE* in a contained namespace, i.e., watch only particular namespace; add the following environment variable `WATCH_NAMESPACE`. This can be easily accompolished with the `downward api` in *Kuberneres* as shown below.

```yaml
- env:
    - name: WATCH_NAMESPACE
        valueFrom:
        fieldRef:
            fieldPath: metadata.namespace
```

Still, *CRE* will require a *service account* which has access to `list` all the `pods` under `core v1` apiGroup and `metrics` apiGroup.

### Sample CRE metrics exported:

```prometheus
# HELP cpu_limit CPU Limit by deployment
# TYPE cpu_limit gauge
cpu_limit{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 1
cpu_limit{container_name="alertmanager-configmap-reload",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 1

# HELP cpu_request Requested CPU by deployment
# TYPE cpu_request gauge
cpu_request{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 0.001
cpu_request{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-nlqqc",status="Running "} 0.001

# HELP current_cpu_usage Current CPU Usage as reported by Metrics API
# TYPE current_cpu_usage gauge
current_cpu_usage{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9"} 0
current_cpu_usage{container_name="alertmanager-configmap-reload",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9"} 0

# HELP current_memory_usage Current CPU Usage as reported by Metrics API
# TYPE current_memory_usage gauge
current_memory_usage{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9"} 1.4168064e+07
current_memory_usage{container_name="alertmanager-configmap-reload",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9"} 1.363968e+06

# HELP memory_limit Memory Limit by deployment
# TYPE memory_limit gauge
memory_limit{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 5.36870912e+08
memory_limit{container_name="alertmanager-configmap-reload",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 1.073741824e+09

# HELP memory_request Requested Memory by deployment
# TYPE memory_request gauge
memory_request{container_name="alertmanager",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 2.68435456e+08
memory_request{container_name="alertmanager-configmap-reload",namespace="default",pod_name="prometheus-alertmanager-74bd9d5867-gmlj9",status="Running "} 5.36870912e+08

# HELP total_pod Total pod count in given space
# TYPE total_pod counter
total_pod{namespace="default"} 1
```

Effortlessly get the Resources' request, limit and current usage by containers in your cluster/namespace.

## Sample Grafana dashboard visualization
The below sample [Grafana](https://grafana.com/) dashboard will show the sample metrics record for 
- Total pods in the namespace over the time
- Total CPU Limit/Request/Usage in the namespace
- Total Memory Limit/Request/Usage in the namespace
- A sample pod's CPU Utilization
- A sample pod's Memory Utilization

![alt](https://github.com/gkarthiks/container-resource-exporter/blob/master/grafana-dashboard.jpeg)

## Docker Image:
The docker image can be found [here <img src="./docker-logo.png" width="40" height="40" align="center"/> .](https://cloud.docker.com/repository/docker/gkarthics/container-resource-exporter)

## Helm Chart:
The helm chart is available [here](./helm-chart) for easy installation.