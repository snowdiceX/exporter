package exporter

// nolint
const (
	KeyMetricAddr   = "metricAddr"
	KeyMetricPath   = "metricPath"
	KeyMetricPrefix = "metricPrefix"
	KeyMetricType   = "metricType"
)

// MetricConfig defined config about metric type
type MetricConfig struct {
	Key    string
	Type   string
	Help   string
	Labels []string
}
