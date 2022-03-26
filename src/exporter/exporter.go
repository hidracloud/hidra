package exporter

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hidracloud/hidra/src/models"
	"github.com/hidracloud/hidra/src/scenarios"
	"github.com/hidracloud/hidra/src/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// mapMutex is a mutex that protects the map
var mapMutex *sync.Mutex

// jobsQueue represent
var jobsQueue chan func(int)

// prometheusLabels contains the labels of the metrics
var prometheusLabels []string

// lastRun is a map of last run time for each sample
var lastRun map[string]time.Time
var inProgress map[string]bool
var lastRunMutex *sync.Mutex

// hidraScenarioStatusVec is a metric vector type which holds hidra status
var hidraScenarioStatusVec *prometheus.GaugeVec
var hidraStepStatusVec *prometheus.GaugeVec

// hidraScenarioElapsedVec is a metric vector type which holds hidra elapsed time
var hidraScenarioElapsedVec *prometheus.HistogramVec
var hidraStepElapsedVec *prometheus.HistogramVec

var hidraScenarioLastRunVec *prometheus.GaugeVec
var hidraScenarioIntervalVec *prometheus.GaugeVec

var hidraCustomMetrics map[string]*prometheus.GaugeVec

func refreshPrometheusMetrics(configFiles []string, buckets []float64) error {
	prometheusLabels = []string{"name", "description", "kind", "config_file"}
	for _, configFile := range configFiles {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}

		sample, err := models.ReadSampleYAML(data)
		if err != nil {
			return err
		}
		for key := range sample.Tags {
			found := false
			for _, label := range prometheusLabels {
				if label == key {
					found = true
					break
				}
			}

			if !found {
				prometheusLabels = append(prometheusLabels, key)
			}
		}
	}

	hidraScenarioStatusVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_sample_metric_status",
		Help: "Status of hidra samples",
	}, prometheusLabels)

	hidraScenarioElapsedVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "hidra_sample_metric_elapsed",
		Help:    "Elapsed time of hidra samples",
		Buckets: buckets,
	}, prometheusLabels)

	stepLabels := []string{}
	stepLabels = append(stepLabels, prometheusLabels...)
	stepLabels = append(stepLabels, "step")

	hidraStepStatusVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_step_metric_status",
		Help: "Status of hidra steps",
	}, stepLabels)

	hidraStepElapsedVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "hidra_step_metric_elapsed",
		Help:    "Elapsed time of hidra steps",
		Buckets: buckets,
	}, stepLabels)

	hidraScenarioLastRunVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_sample_metric_last_run",
		Help: "Last run time of hidra samples",
	}, prometheusLabels)

	hidraScenarioIntervalVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_sample_metric_interval",
		Help: "Interval of hidra samples",
	}, prometheusLabels)

	hidraCustomMetrics = make(map[string]*prometheus.GaugeVec)

	// Restart prometheus
	prometheus.MustRegister(hidraScenarioStatusVec)
	prometheus.MustRegister(hidraScenarioElapsedVec)
	prometheus.MustRegister(hidraStepStatusVec)
	prometheus.MustRegister(hidraStepElapsedVec)
	prometheus.MustRegister(hidraScenarioLastRunVec)
	prometheus.MustRegister(hidraScenarioIntervalVec)

	return nil
}

func readLabels(sample *models.Sample, configFile string) []string {
	labels := []string{}

	labels = append(labels, sample.Name)
	labels = append(labels, sample.Description)
	labels = append(labels, sample.Scenario.Kind)
	labels = append(labels, "")

	for _, label := range prometheusLabels[4:] {
		foundVal := ""
		for key, val := range sample.Tags {
			if key == label {
				foundVal = val
				break
			}
		}

		labels = append(labels, foundVal)
	}

	return labels
}

func createCustomMetricIfDontExists(metric *models.Metric) {
	mapMutex.Lock()
	defer mapMutex.Unlock()
	if _, ok := hidraCustomMetrics[metric.Name]; !ok {
		metricLabels := []string{}
		metricLabels = append(metricLabels, prometheusLabels...)
		metricLabels = append(metricLabels, "step")

		for label := range metric.Labels {
			metricLabels = append(metricLabels, label)
		}

		hidraCustomMetrics[metric.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("hidra_custom_%s", metric.Name),
			Help: metric.Description,
		}, metricLabels)

		prometheus.MustRegister(hidraCustomMetrics[metric.Name])
	}
}

func runOneScenario(ctx context.Context, sample *models.Sample, configFile string) {
	workerID := ctx.Value(utils.CustomContextKey("workerID")).(int)
	log.Println("["+strconv.Itoa(workerID)+"] Running scenario:", sample.Name, "with description:", sample.Description)
	m, err := scenarios.RunScenario(ctx, sample.Scenario, sample.Name, sample.Description)

	if err != nil {
		log.Println("Error running scenario:", err)
		return
	}

	status := 0

	if m.Error == nil {
		status = 1
	}

	labels := readLabels(sample, configFile)
	hidraScenarioStatusVec.WithLabelValues(labels...).Set(float64(status))
	hidraScenarioElapsedVec.WithLabelValues(labels...).Observe(float64(m.EndDate.UnixMilli() - m.StartDate.UnixMilli()))

	for _, step := range m.StepResults {
		stepLabels := append(labels, step.Step.Type)

		for _, metric := range step.Metrics {
			// Check if custom metric exists
			createCustomMetricIfDontExists(&metric)

			metricLabels := []string{}
			metricLabels = append(metricLabels, stepLabels...)

			for _, label := range metric.Labels {
				metricLabels = append(metricLabels, label)
			}

			mapMutex.Lock()
			hidraCustomMetrics[metric.Name].WithLabelValues(metricLabels...).Set(float64(metric.Value))
			mapMutex.Unlock()
		}

		hidraStepElapsedVec.WithLabelValues(stepLabels...).Observe(float64(step.EndDate.UnixMilli() - step.StartDate.UnixMilli()))
	}

	hidraScenarioLastRunVec.WithLabelValues(labels...).Set(float64(time.Now().UnixMilli()))
	hidraScenarioIntervalVec.WithLabelValues(labels...).Set(float64(sample.ScrapeInterval))
}

func runSample(ctx context.Context, configFiles []string, maxExecutors int) {
	log.Println("Calculating samples to run")

	newSamples := 0
	for _, configFile := range configFiles {
		data, _ := ioutil.ReadFile(configFile)
		sample, _ := models.ReadSampleYAML(data)

		// Check last run
		lastRunTime, ok := lastRun[sample.Name]
		if !ok {
			lastRunTime = time.Unix(0, 0)
			lastRun[sample.Name] = lastRunTime
		}

		if _, ok = inProgress[sample.Name]; !ok {
			inProgress[sample.Name] = false
		}

		// Check if it's time to run
		if time.Since(lastRunTime) < sample.ScrapeInterval || inProgress[sample.Name] {
			continue
		}

		newSamples++

		lastRunMutex.Lock()
		inProgress[sample.Name] = true
		lastRunMutex.Unlock()

		jobsQueue <- func(workerID int) {
			runOneScenario(context.WithValue(ctx, utils.CustomContextKey("workerID"), workerID), sample, configFile)
			lastRunMutex.Lock()
			randomOffset := time.Duration(rand.Intn(int(sample.ScrapeInterval.Seconds()))) * time.Second
			lastRun[sample.Name] = time.Now().Add(randomOffset)
			inProgress[sample.Name] = false
			lastRunMutex.Unlock()
		}
	}

	log.Println("Running", newSamples, "samples")
}

func checkDuplicatedSamples(configFiles []string) {
	log.Println("Checking duplicated samples")

	errors := 0
	processedSample := make(map[string]bool)
	for _, configFile := range configFiles {
		data, _ := ioutil.ReadFile(configFile)
		sample, _ := models.ReadSampleYAML(data)

		if _, ok := processedSample[sample.Name]; ok {
			log.Println("Duplicated sample:", sample.Name, "in", configFile)

			errors++
		}

		processedSample[sample.Name] = true
	}

	if errors > 0 {
		log.Fatal("Found duplicated samples")
	}
}

func metricsRecord(ctx context.Context, confPath string, maxExecutor int, buckets []float64) {
	configFiles, err := utils.AutoDiscoverYML(confPath)
	if err != nil {
		panic(err)
	}

	checkDuplicatedSamples(configFiles)

	log.Println("Reloading prometheus metrics")
	err = refreshPrometheusMetrics(configFiles, buckets)
	if err != nil {
		panic(err)
	}
	log.Println("Prometheus metrics reloaded")

	lastRun = make(map[string]time.Time)
	inProgress = make(map[string]bool)
	lastRunMutex = &sync.Mutex{}

	createWorkers(maxExecutor, len(configFiles))

	go func() {
		for {
			runSample(ctx, configFiles, maxExecutor)
			time.Sleep(2 * time.Second)
		}
	}()
}

func createWorkers(maxExecutor, possibleJobs int) {
	log.Printf("Creating %d workers", maxExecutor)
	jobsQueue = make(chan func(int), possibleJobs)
	mapMutex = &sync.Mutex{}

	for i := 0; i < maxExecutor; i++ {
		go func(workerID int) {
			log.Println("Initializing worker", workerID)
			for {
				job := <-jobsQueue
				log.Println("[" + strconv.Itoa(workerID) + "] running job")
				job(workerID)
				log.Println("[" + strconv.Itoa(workerID) + "] run job successfully")
			}
		}(i)
	}
}

// Run starts the metrics recorder
func Run(ctx context.Context, wg *sync.WaitGroup, confPath string, maxExecutor, port int, buckets []float64) {
	log.Println("Starting hidra in exporter mode")

	// Start fetching metrics
	metricsRecord(ctx, confPath, maxExecutor, buckets)

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
	wg.Done()
}
