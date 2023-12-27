package exporter

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hidracloud/hidra/v3/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/runner"
	"github.com/hidracloud/hidra/v3/internal/utils"
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

	// sampleRunningTimeLock is the sample running time lock
	sampleRunningTimeMutex *sync.RWMutex

	// lastSchedulerRun is the last scheduler run
	lastRun map[string]time.Time

	// lastRunMutex is the last run mutex
	lastRunMutex *sync.RWMutex

	// prometheusMetricStore is the prometheus metric store
	prometheusMetricStore = make(map[string]*prometheus.GaugeVec)

	// prometheusMetricStoreMutex is the prometheus metric store mutex
	prometheusMetricStoreMutex = &sync.RWMutex{}

	// prometheusStatusMetricStore is the prometheus status metric store
	prometheusStatusMetric *prometheus.GaugeVec

	// prometheusLastUpdate is the last time the metrics were updated
	prometheusLastUpdate *prometheus.GaugeVec

	// enableUsage is the flag to enable the usage
	enableUsage = false

	// prometheusTimeToRun is the time to run the sample
	prometheusTimeToRun *prometheus.GaugeVec

	// toBePurged is the to be purged
	toBePurged = make(map[string]struct {
		Labels           map[string]string
		PurgeAt          time.Time
		PrometheusMetric *prometheus.GaugeVec
	})

	// toBePurgedMutex is the to be purged mutex
	toBePurgedMutex = &sync.Mutex{}
)

// InitializeWorker initializes the worker
func InitializeWorker(config *config.ExporterConfig) {
	runningTime = &atomic.Uint64{}
	sampleRunningTime = make(map[string]*atomic.Uint64)
	sampleRunningTimeMutex = &sync.RWMutex{}
	lastRun = make(map[string]time.Time)
}

// add2PurgeList adds to purge list
func add2PurgeList(purgeLabels prometheus.Labels, purgeTime time.Time, prometheusMetric *prometheus.GaugeVec) {
	log.Debug("Adding metric to purge list", purgeLabels)
	toBePurgedMutex.Lock()
	defer toBePurgedMutex.Unlock()
	toBePurged[utils.Map2Hash(purgeLabels)] = struct {
		Labels           map[string]string
		PurgeAt          time.Time
		PrometheusMetric *prometheus.GaugeVec
	}{
		Labels:           purgeLabels,
		PurgeAt:          purgeTime,
		PrometheusMetric: prometheusMetric,
	}
}

// purgeMetrics purges the metrics
func purgeMetrics() {
	for {
		toBePurgedMutex.Lock()

		log.Debug("Purging metrics")

		for hash, metric := range toBePurged {
			if time.Now().After(metric.PurgeAt) {
				log.Debug("Purging metric", metric.Labels)
				metric.PrometheusMetric.DeletePartialMatch(metric.Labels)
				delete(toBePurged, hash)
			}
		}

		toBePurgedMutex.Unlock()

		time.Sleep(1 * time.Minute)
	}

}

// addMetric2PurgeListIfNeeeded adds the metric to purge list if needed
func addMetric2PurgeListIfNeeeded(metric *metrics.Metric, sample *config.SampleConfig) {
	if metric.Purge {
		prometheusMetric := initializePrometheusMetrics(metric)
		labels := createLabels(metric, sample)

		// Get only labels present on metric.PurgeLabels
		purgeLabels := prometheus.Labels{}

		for _, label := range metric.PurgeLabels {
			if value, ok := labels[label]; ok {
				purgeLabels[label] = value
			}
		}

		purgeTime := time.Now().Add(metric.PurgeAfter)

		add2PurgeList(purgeLabels, purgeTime, prometheusMetric)
	}
}

// createMetrics creates the metrics
func createMetrics(metric *metrics.Metric, sample *config.SampleConfig) {
	prometheusMetric := initializePrometheusMetrics(metric)
	labels := createLabels(metric, sample)

	prometheusMetric.With(labels).Set(metric.Value)
}

// updateMetrics updates the metrics
func updateMetrics(allMetrics []*metrics.Metric, sample *config.SampleConfig, startTime time.Time, err error) {
	// Purge metrics if metric.Purge is true
	for _, metric := range allMetrics {
		addMetric2PurgeListIfNeeeded(metric, sample)
		createMetrics(metric, sample)
	}

	statusLabels := createLabelsForStatus(sample)

	if err != nil {
		prometheusStatusMetric.With(statusLabels).Set(0)
	} else {
		prometheusStatusMetric.With(statusLabels).Set(1)
	}

	if enableUsage {
		prometheusTimeToRun.With(statusLabels).Set(float64(time.Since(startTime).Milliseconds()))
	}

	prometheusLastUpdate.With(statusLabels).Set(float64(time.Now().Unix()))
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
	prometheusMetricStoreMutex.RLock()
	defer prometheusMetricStoreMutex.RUnlock()
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

// RunSampleWithTimeout runs the sample with timeout
func RunSampleWithTimeout(ctx context.Context, sample *config.SampleConfig, timeout time.Duration) *runner.RunnerResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return runner.RunSample(ctx, sample)
}

// RunOneWorker runs one worker
func RunOneWorker(worker int, config *config.ExporterConfig) {
	for {
		sample := <-samplesJobs
		startTime := time.Now()
		log.Debugf("Running sample %s, with description %s from worker %d", sample.Name, sample.Description, worker)

		// Run the sample
		ctx, cancel := context.WithCancel(context.Background())
		result := RunSampleWithTimeout(ctx, sample, sample.Timeout)

		// Update the metrics
		updateMetrics(result.Metrics, sample, startTime, result.Error)

		runningTime.Add(uint64(time.Since(startTime).Milliseconds()))

		sampleRunningTimeMutex.Lock()
		if _, ok := sampleRunningTime[sample.Name]; !ok {
			sampleRunningTime[sample.Name] = &atomic.Uint64{}
		}

		randomOffset := time.Duration(rand.Intn(int(sample.Interval.Seconds()))) * time.Second
		sampleRunningTime[sample.Name].Add(uint64(time.Since(startTime).Milliseconds()))
		sampleRunningTimeMutex.Unlock()

		lastRunMutex.Lock()
		lastRun[sample.Name] = time.Now().Add(randomOffset)
		lastRunMutex.Unlock()

		if time.Since(startTime) > 30*time.Second {
			log.Warnf("Sample %s took more than a minute from worker %d", sample.Name, worker)
		}

		cancel()

		time.Sleep(config.WorkerConfig.SleepBetweenJobs)
	}
}

// runBackgroundMetricsTask runs the background metrics task
func runBackgroundMetricsTask() {
	for {
		backgroundTask := runner.GetNextBackgroundTask()

		for backgroundTask != nil {
			log.Debug("Running background task")
			allMetrics, sample, err := backgroundTask()

			if err != nil {
				log.Debug("Error getting background task metrics", err)
			}

			if sample != nil {
				for _, metric := range allMetrics {
					addMetric2PurgeListIfNeeeded(metric, sample)
					createMetrics(metric, sample)
				}
			}

			backgroundTask = runner.GetNextBackgroundTask()
		}
		time.Sleep(10 * time.Second)
	}
}

// RunWorkers runs the workers
func RunWorkers(cnf *config.ExporterConfig) {

	log.Infof("Initializing %d workers...", cnf.WorkerConfig.ParallelJobs)

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

	if cnf.UsageConfig.Enabled {
		enableUsage = true

		prometheusTimeToRun = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:      "time_to_run",
				Help:      "Hidra time to run",
				Namespace: "hidra",
			},
			sampleCommonTags,
		)

		prometheus.MustRegister(prometheusTimeToRun)
	}

	go purgeMetrics()

	if !cnf.WorkerConfig.DisableBGTasks {
		go runBackgroundMetricsTask()
	} else {
		runner.DisableBackgroundTask = true
	}

	for i := 0; i < cnf.WorkerConfig.ParallelJobs; i++ {
		go RunOneWorker(i, cnf)
	}

}
