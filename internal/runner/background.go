package runner

import (
	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
)

var (
	// BackgroundTask are task that must be run in background
	BackgroundTask = []func() ([]*metrics.Metric, *config.SampleConfig, error){}
)

// RegisterBackgroundTask register a background task
func RegisterBackgroundTask(f func() ([]*metrics.Metric, *config.SampleConfig, error)) {
	BackgroundTask = append(BackgroundTask, f)
}

// GetNextBackgroundTask return the next background task
func GetNextBackgroundTask() func() ([]*metrics.Metric, *config.SampleConfig, error) {
	if len(BackgroundTask) == 0 {
		return nil
	}
	f := BackgroundTask[0]
	BackgroundTask = BackgroundTask[1:]
	return f
}

// init register all background task
func init() {
	BackgroundTask = make([]func() ([]*metrics.Metric, *config.SampleConfig, error), 0)
}
