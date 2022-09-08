package metrics

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
}
