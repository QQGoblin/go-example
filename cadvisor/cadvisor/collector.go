package cadvisor

import (
	"github.com/google/cadvisor/container"
	cadvisormetrics "github.com/google/cadvisor/container"
	info "github.com/google/cadvisor/info/v1"
	cadvisorv2 "github.com/google/cadvisor/info/v2"
	"github.com/google/cadvisor/manager"
	"github.com/google/cadvisor/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

var (
	versionInfoDesc = prometheus.NewDesc("cadvisor_version_info", "A metric with a constant '1' value labeled by kernel version, OS version, docker version, cadvisor version & cadvisor revision.", []string{"kernelVersion", "osVersion", "dockerVersion", "cadvisorVersion", "cadvisorRevision"}, nil)
	startTimeDesc   = prometheus.NewDesc("container_start_time_seconds", "Start time of the container since unix epoch in seconds.", nil, nil)
	cpuPeriodDesc   = prometheus.NewDesc("container_spec_cpu_period", "CPU period of the container.", nil, nil)
	cpuQuotaDesc    = prometheus.NewDesc("container_spec_cpu_quota", "CPU quota of the container.", nil, nil)
	cpuSharesDesc   = prometheus.NewDesc("container_spec_cpu_shares", "CPU share of the container.", nil, nil)

	defaultContainerMetrics = []containerMetric{
		ContainerFsReadsBytesTotalMetric,
		ContainerFsWritesBytesTotalMetric,
		ContainerNetworkReceiveBytesTotalMetric,
		ContainerNetworkTransmitBytesTotalMetric,
		ContainerCpuUsageSecondsTotalMetric,
		ContainerMemoryUsageBytesMetric,
	}
	defaultIncludedMetrics = cadvisormetrics.MetricSet{
		cadvisormetrics.CpuUsageMetrics:     struct{}{},
		cadvisormetrics.MemoryUsageMetrics:  struct{}{},
		cadvisormetrics.DiskIOMetrics:       struct{}{},
		cadvisormetrics.NetworkUsageMetrics: struct{}{},
	}
)

type ContainerLabelsFunc func(*info.ContainerInfo) map[string]string

func NewCollector(m manager.Manager, f ContainerLabelsFunc) Collector {

	return Collector{
		infoProvider: m,
		errors: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "container",
			Name:      "scrape_error",
			Help:      "1 if there was an error while getting container metrics, 0 otherwise",
		}),
		containerMetrics:    defaultContainerMetrics,
		containerLabelsFunc: f,
		includedMetrics:     defaultIncludedMetrics,
		opts: cadvisorv2.RequestOptions{
			IdType:    cadvisorv2.TypeName,
			Count:     1,
			Recursive: true,
		},
	}
}

// Collector implements prometheus.Collector.
type Collector struct {
	infoProvider        manager.Manager
	errors              prometheus.Gauge
	containerMetrics    []containerMetric
	containerLabelsFunc ContainerLabelsFunc
	includedMetrics     container.MetricSet
	opts                cadvisorv2.RequestOptions
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.errors.Describe(ch)
	for _, cm := range c.containerMetrics {
		ch <- cm.desc([]string{})
	}
	ch <- startTimeDesc
	ch <- cpuPeriodDesc
	ch <- cpuQuotaDesc
	ch <- cpuSharesDesc
	ch <- versionInfoDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.errors.Set(0)
	c.collectVersionInfo(ch)
	c.collectContainersInfo(ch)
	c.errors.Collect(ch)
}

func (c *Collector) collectVersionInfo(ch chan<- prometheus.Metric) {
	versionInfo, err := c.infoProvider.GetVersionInfo()
	if err != nil {
		c.errors.Set(1)
		klog.Warningf("Couldn't get version info: %s", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(versionInfoDesc, prometheus.GaugeValue, 1, []string{versionInfo.KernelVersion, versionInfo.ContainerOsVersion, versionInfo.DockerVersion, versionInfo.CadvisorVersion, versionInfo.CadvisorRevision}...)
}

func (c *Collector) collectContainersInfo(ch chan<- prometheus.Metric) {
	containers, err := c.infoProvider.GetRequestedContainersInfo("/", c.opts)
	if err != nil {
		c.errors.Set(1)
		klog.Warningf("Couldn't get containers: %s", err)
		return
	}
	rawLabels := map[string]struct{}{}
	for _, container := range containers {
		for l := range c.containerLabelsFunc(container) {
			rawLabels[l] = struct{}{}
		}
	}

	for _, cont := range containers {

		values := make([]string, 0, len(rawLabels))
		labels := make([]string, 0, len(rawLabels))
		containerLabels := c.containerLabelsFunc(cont)
		for l := range rawLabels {
			duplicate := false
			sl := sanitizeLabelName(l)
			for _, x := range labels {
				if sl == x {
					duplicate = true
					break
				}
			}
			if !duplicate {
				labels = append(labels, sl)
				values = append(values, containerLabels[l])
			}
		}

		// Container spec
		if cont.Spec.HasMemory {
			desc := prometheus.NewDesc("container_spec_memory_limit_bytes", "Memory limit for the container.", labels, nil)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, specMemoryValue(cont.Spec.Memory.Limit), values...)
		}

		// Now for the actual metrics
		if len(cont.Stats) == 0 {
			continue
		}
		stats := cont.Stats[0]
		for _, cm := range c.containerMetrics {
			if cm.condition != nil && !cm.condition(cont.Spec) {
				continue
			}
			desc := cm.desc(labels)
			for _, metricValue := range cm.getValues(stats) {
				ch <- prometheus.NewMetricWithTimestamp(
					metricValue.timestamp,
					prometheus.MustNewConstMetric(desc, cm.valueType, float64(metricValue.value), append(values, metricValue.labels...)...),
				)
			}
		}
	}
}

func containerLabels(c *info.ContainerInfo) map[string]string {
	//var name, image, podName, namespace, containerName string
	var name, podName, namespace, containerName string
	if len(c.Aliases) > 0 {
		name = c.Aliases[0]
	}
	//image = c.Spec.Image
	if v, ok := c.Spec.Labels[KubernetesPodNameLabel]; ok {
		podName = v
	}
	if v, ok := c.Spec.Labels[KubernetesPodNamespaceLabel]; ok {
		namespace = v
	}
	if v, ok := c.Spec.Labels[KubernetesContainerNameLabel]; ok {
		containerName = v
	}

	set := map[string]string{
		//metrics.LabelID:   c.Name,
		metrics.LabelName: name,
		//metrics.LabelImage: image,
		"pod":       podName,
		"namespace": namespace,
		"container": containerName,
	}

	return set
}
