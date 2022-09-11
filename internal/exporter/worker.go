package exporter

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/plugins"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	// samplesJobs is the samples jobs
	samplesJobs chan *config.SampleConfig

	// runningTime is the running time
	runningTime *atomic.Uint64

	// runningSamples is the running samples
	sampleRunningTime map[string]*atomic.Uint64

	// lastSchedulerRun is the last scheduler run
	lastRun map[string]time.Time

	// lastRunMutex is the mutex to protect the last scheduler run
	lastRunMutex *sync.RWMutex

	// prometheusMetricStore is the prometheus metric store
	prometheusMetricStore = make(map[string]*prometheus.GaugeVec)

	// prometheusStatusMetricStore is the prometheus status metric store
	prometheusStatusMetric *prometheus.GaugeVec

	// prometheusLastUpdate is the last time the metrics were updated
	prometheusLastUpdate *prometheus.GaugeVec
)

// InitializeWorker initializes the worker
func InitializeWorker(config *config.ExporterConfig) {
	runningTime = &atomic.Uint64{}
	sampleRunningTime = make(map[string]*atomic.Uint64)
	lastRun = make(map[string]time.Time)
	lastRunMutex = &sync.RWMutex{}
}

// updateMetrics updates the metrics
func updateMetrics(allMetrics []*metrics.Metric, sample *config.SampleConfig, err error) {
	for _, metric := range allMetrics {
		prometheusMetric := initializePrometheusMetrics(metric)
		labels := createLabels(metric, sample)

		prometheusMetric.With(labels).Set(metric.Value)

		statusLabels := createLabelsForStatus(sample)

		if err != nil {
			prometheusStatusMetric.With(statusLabels).Set(0)
		} else {
			prometheusStatusMetric.With(statusLabels).Set(1)
		}

		prometheusLastUpdate.With(statusLabels).Set(float64(time.Now().Unix()))
	}
}

// createLabelsForStatus creates the labels for status
func createLabelsForStatus(sample *config.SampleConfig) prometheus.Labels {
	return createLabels(&metrics.Metric{}, sample)
}

// createLabels creates the labels
func createLabels(metric *metrics.Metric, sample *config.SampleConfig) prometheus.Labels {
	metricLabels := make([]string, 0)

	for label := range metric.Labels {
		metricLabels = append(metricLabels, label)
	}

	metricLabels = append(metricLabels, sampleCommonTags...)

	labels := prometheus.Labels{}

	for _, label := range metricLabels {
		labels[label] = ""
		if _, ok := metric.Labels[label]; ok {
			labels[label] = metric.Labels[label]
		} else if _, ok := sample.Tags[label]; ok {
			labels[label] = sample.Tags[label]
		}
	}

	pluginList := make(map[string]bool, 0)
	for _, step := range sample.Steps {
		if step.Plugin != "" {
			pluginList[step.Plugin] = true
		}
	}

	allPlugins := make([]string, 0)

	for plugin := range pluginList {
		allPlugins = append(allPlugins, plugin)
	}

	sort.Strings(allPlugins)

	labels["sample_name"] = sample.Name
	labels["plugins"] = strings.Join(allPlugins, ",")
	labels["description"] = sample.Description
	return labels
}

// initializePrometheusMetrics initializes the prometheus metrics
func initializePrometheusMetrics(metric *metrics.Metric) *prometheus.GaugeVec {
	if _, ok := prometheusMetricStore[metric.Name]; !ok {
		metricLabels := make([]string, 0)

		for label := range metric.Labels {
			metricLabels = append(metricLabels, label)
		}

		metricLabels = append(metricLabels, sampleCommonTags...)

		prometheusMetricStore[metric.Name] = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      metric.Name,
				Help:      metric.Description,
				Namespace: "hidra",
			},
			metricLabels,
		)
		prometheus.MustRegister(prometheusMetricStore[metric.Name])
	}

	return prometheusMetricStore[metric.Name]
}

// RunWorkers runs the workers
func RunWorkers(cnf *config.ExporterConfig) {

	log.Debugf("Initializing %d workers...", cnf.WorkerConfig.ParallelJobs)

	samplesJobs = make(chan *config.SampleConfig, cnf.WorkerConfig.MaxQueueSize)

	prometheusStatusMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "sample_status",
			Help:      "Hidra sample status",
			Namespace: "hidra",
		},
		sampleCommonTags,
	)

	prometheusLastUpdate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "last_update",
			Help:      "Hidra last update",
			Namespace: "hidra",
		},
		sampleCommonTags,
	)

	prometheus.MustRegister(prometheusLastUpdate)
	prometheus.MustRegister(prometheusStatusMetric)

	for i := 0; i < cnf.WorkerConfig.ParallelJobs; i++ {
		go func(worker int) {
			for {
				sample := <-samplesJobs
				startTime := time.Now()
				log.Debugf("Running sample %s, with description %s from worker %d", sample.Name, sample.Description, worker)

				// Run the sample
				ctx := context.Background()
				_, allMetrics, err := plugins.RunSample(ctx, sample)

				// Update the metrics
				updateMetrics(allMetrics, sample, err)

				runningTime.Add(uint64(time.Since(startTime).Milliseconds()))

				lastRunMutex.Lock()

				if _, ok := sampleRunningTime[sample.Name]; !ok {
					sampleRunningTime[sample.Name] = &atomic.Uint64{}
				}

				randomOffset := time.Duration(rand.Intn(int(sample.Interval.Seconds()))) * time.Second
				sampleRunningTime[sample.Name].Add(uint64(time.Since(startTime).Milliseconds()))
				lastRun[sample.Name] = time.Now().Add(randomOffset)
				lastRunMutex.Unlock()

				time.Sleep(cnf.WorkerConfig.SleepBetweenJobs)
			}
		}(i)
	}
}
