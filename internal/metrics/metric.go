package metrics

import "time"

// Metric represents a metric.
type Metric struct {
	// Name is the name of the metric.
	Name string
	// Value is the value of the metric.
	Value float64
	// Labels is the tags of the metric.
	Labels map[string]string
	// Description is the description of the metric.
	Description string
	// Purge is the purge of the metric.
	Purge bool `default:"false"`
	// PurgeLabels is the purge labels of the metric.
	PurgeLabels []string
	// PurgeAfter is the purge after of the metric.
	PurgeAfter time.Duration
}

// MetricsToMap converts metrics to map.
func MetricsToMap(metrics []*Metric) map[string]float64 {
	result := make(map[string]float64)
	for _, metric := range metrics {
		result[metric.Name] = metric.Value
	}
	return result
}
