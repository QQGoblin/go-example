package cadvisor

import (
	info "github.com/google/cadvisor/info/v1"
	"regexp"
	"time"
)

const (
	maxMemorySize = uint64(1 << 62)
)

func specMemoryValue(v uint64) float64 {
	if v > maxMemorySize {
		return 0
	}
	return float64(v)
}

var invalidNameCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizeLabelName replaces anything that doesn't match
// client_label.LabelNameRE with an underscore.
func sanitizeLabelName(name string) string {
	return invalidNameCharRE.ReplaceAllString(name, "_")
}

func ioValues(ioStats []info.PerDiskStats, ioType string, ioValueFn func(uint64) float64, fsStats []info.FsStats, valueFn func(*info.FsStats) float64, timestamp time.Time) metricValues {

	values := make(metricValues, 0, len(ioStats)+len(fsStats))
	for _, stat := range ioStats {
		values = append(values, metricValue{
			value:     ioValueFn(stat.Stats[ioType]),
			labels:    []string{stat.Device},
			timestamp: timestamp,
		})
	}
	for _, stat := range fsStats {
		values = append(values, metricValue{
			value:     valueFn(&stat),
			labels:    []string{stat.Device},
			timestamp: timestamp,
		})
	}
	return values
}
func asFloat64(v uint64) float64 { return float64(v) }
