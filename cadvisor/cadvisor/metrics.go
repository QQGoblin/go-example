package cadvisor

import (
	"fmt"
	info "github.com/google/cadvisor/info/v1"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type metricValue struct {
	value     float64
	labels    []string
	timestamp time.Time
}

type metricValues []metricValue

type containerMetric struct {
	name        string
	help        string
	valueType   prometheus.ValueType
	extraLabels []string
	condition   func(s info.ContainerSpec) bool
	getValues   func(s *info.ContainerStats) metricValues
}

func (cm *containerMetric) desc(baseLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(cm.name, cm.help, append(baseLabels, cm.extraLabels...), nil)
}

var (
	ContainerFsWritesBytesTotalMetric = containerMetric{
		name:        "container_fs_writes_bytes_total",
		help:        "Cumulative count of bytes written",
		valueType:   prometheus.CounterValue,
		extraLabels: []string{"device"},
		getValues: func(s *info.ContainerStats) metricValues {
			return ioValues(
				s.DiskIo.IoServiceBytes, "Write", asFloat64,
				nil, nil,
				s.Timestamp,
			)
		},
	}
	ContainerFsReadsBytesTotalMetric = containerMetric{
		name:        "container_fs_reads_bytes_total",
		help:        "Cumulative count of bytes read",
		valueType:   prometheus.CounterValue,
		extraLabels: []string{"device"},
		getValues: func(s *info.ContainerStats) metricValues {
			return ioValues(
				s.DiskIo.IoServiceBytes, "Read", asFloat64,
				nil, nil,
				s.Timestamp,
			)
		},
	}
	ContainerNetworkReceiveBytesTotalMetric = containerMetric{
		name:        "container_network_receive_bytes_total",
		help:        "Cumulative count of bytes received",
		valueType:   prometheus.CounterValue,
		extraLabels: []string{"interface"},
		getValues: func(s *info.ContainerStats) metricValues {
			values := make(metricValues, 0, len(s.Network.Interfaces))
			for _, value := range s.Network.Interfaces {
				values = append(values, metricValue{
					value:     float64(value.RxBytes),
					labels:    []string{value.Name},
					timestamp: s.Timestamp,
				})
			}
			return values
		},
	}
	ContainerNetworkTransmitBytesTotalMetric = containerMetric{
		name:        "container_network_transmit_bytes_total",
		help:        "Cumulative count of bytes transmitted",
		valueType:   prometheus.CounterValue,
		extraLabels: []string{"interface"},
		getValues: func(s *info.ContainerStats) metricValues {
			values := make(metricValues, 0, len(s.Network.Interfaces))
			for _, value := range s.Network.Interfaces {
				values = append(values, metricValue{
					value:     float64(value.TxBytes),
					labels:    []string{value.Name},
					timestamp: s.Timestamp,
				})
			}
			return values
		},
	}
	ContainerMemoryUsageBytesMetric = containerMetric{
		name:      "container_memory_usage_bytes",
		help:      "Current memory usage in bytes, including all memory regardless of when it was accessed",
		valueType: prometheus.GaugeValue,
		getValues: func(s *info.ContainerStats) metricValues {
			return metricValues{{value: float64(s.Memory.Usage), timestamp: s.Timestamp}}
		},
	}
	ContainerCpuUsageSecondsTotalMetric = containerMetric{
		name:        "container_cpu_usage_seconds_total",
		help:        "Cumulative cpu time consumed in seconds.",
		valueType:   prometheus.CounterValue,
		extraLabels: []string{"cpu"},
		getValues: func(s *info.ContainerStats) metricValues {
			if len(s.Cpu.Usage.PerCpu) == 0 {
				if s.Cpu.Usage.Total > 0 {
					return metricValues{{
						value:     float64(s.Cpu.Usage.Total) / float64(time.Second),
						labels:    []string{"total"},
						timestamp: s.Timestamp,
					}}
				}
			}
			values := make(metricValues, 0, len(s.Cpu.Usage.PerCpu))
			for i, value := range s.Cpu.Usage.PerCpu {
				if value > 0 {
					values = append(values, metricValue{
						value:     float64(value) / float64(time.Second),
						labels:    []string{fmt.Sprintf("cpu%02d", i)},
						timestamp: s.Timestamp,
					})
				}
			}
			return values
		},
	}
)
