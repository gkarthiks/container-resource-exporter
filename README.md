# container-resource-exporter
![Build Type](https://img.shields.io/docker/cloud/automated/gkarthics/container-resource-exporter.svg)
![Build Status](https://img.shields.io/docker/cloud/build/gkarthics/container-resource-exporter.svg)
![License](https://img.shields.io/github/license/gkarthiks/container-resource-exporter.svg)
![Release](https://img.shields.io/github/tag-date/gkarthiks/container-resource-exporter.svg?color=Orange&label=Latest%20Release)

Container Resource Exporter (CRE) is a simple metrics expoerter which will export the following data in the [Prometheus](https://prometheus.io/) format. This can be utilized to trigger a pro-active alert from the the [Prometheus Alert Manager](https://prometheus.io/docs/alerting/alertmanager).

### Sample CRE metrics exported:

```
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
