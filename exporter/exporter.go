package exporter

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
	"github.com/hidracloud/hidra/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// prometheusLabels contains the labels of the metrics
var prometheusLabels []string

// lastRun is a map of last run time for each sample
var lastRun map[string]time.Time

// hidraScenarioStatusVec is a metric vector type which holds hidra status
var hidraScenarioStatusVec *prometheus.GaugeVec
var hidraStepStatusVec *prometheus.GaugeVec

// hidraScenarioElapsedVec is a metric vector type which holds hidra elapsed time
var hidraScenarioElapsedVec *prometheus.GaugeVec
var hidraStepElapsedVec *prometheus.GaugeVec

var hidraScenarioLastRunVec *prometheus.GaugeVec
var hidraScenarioIntervalVec *prometheus.GaugeVec

var hidraCustomMetrics map[string]*prometheus.GaugeVec

func refreshPrometheusMetrics(configFiles []string) error {
	// hidraScenarioStatusVec := prometheus.NewGaugeVec()
	prometheusLabels = []string{"name", "description", "kind", "config_file"}
	for _, configFile := range configFiles {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}

		sample, err := models.ReadScenariosYAML(data)
		if err != nil {
			return err
		}
		for key, _ := range sample.Tags {
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

	hidraScenarioElapsedVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_sample_metric_elapsed",
		Help: "Elapsed time of hidra samples",
	}, prometheusLabels)

	stepLabels := []string{}
	stepLabels = append(stepLabels, prometheusLabels...)
	stepLabels = append(stepLabels, "step")

	hidraStepStatusVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_step_metric_status",
		Help: "Status of hidra steps",
	}, stepLabels)

	hidraStepElapsedVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "hidra_step_metric_elapsed",
		Help: "Elapsed time of hidra steps",
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

func readLabels(sample *models.Scenarios, configFile string) []string {
	labels := []string{}

	labels = append(labels, sample.Name)
	labels = append(labels, sample.Description)
	labels = append(labels, sample.Scenario.Kind)
	labels = append(labels, configFile)

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
	if _, ok := hidraCustomMetrics[metric.Name]; !ok {
		metricLabels := []string{}
		metricLabels = append(metricLabels, prometheusLabels...)
		metricLabels = append(metricLabels, "step")

		for label, _ := range metric.Labels {
			metricLabels = append(metricLabels, label)
		}

		hidraCustomMetrics[metric.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("hidra_custom_%s", metric.Name),
			Help: metric.Description,
		}, metricLabels)

		prometheus.MustRegister(hidraCustomMetrics[metric.Name])
	}
}

func runOneScenario(sample *models.Scenarios, configFile string) {
	m := scenarios.RunScenario(sample.Scenario, sample.Name, sample.Description)

	status := 0

	if m.Error == nil {
		status = 1
	}

	labels := readLabels(sample, configFile)
	hidraScenarioStatusVec.WithLabelValues(labels...).Set(float64(status))
	hidraScenarioElapsedVec.WithLabelValues(labels...).Set(float64(m.EndDate.UnixMilli() - m.StartDate.UnixMilli()))

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

			hidraCustomMetrics[metric.Name].WithLabelValues(metricLabels...).Set(float64(metric.Value))
		}

		hidraStepElapsedVec.WithLabelValues(stepLabels...).Set(float64(step.EndDate.UnixMilli() - step.StartDate.UnixMilli()))
	}

	hidraScenarioLastRunVec.WithLabelValues(labels...).Set(float64(time.Now().UnixMilli()))
	hidraScenarioIntervalVec.WithLabelValues(labels...).Set(float64(sample.ScrapeInterval))
}

func runScenarios(configFiles []string, maxExecutors int) {
	toRun := make([]func(), 0)

	for _, configFile := range configFiles {
		data, _ := ioutil.ReadFile(configFile)
		sample, _ := models.ReadScenariosYAML(data)

		// Check last run
		lastRunTime, ok := lastRun[sample.Name]
		if !ok {
			lastRunTime = time.UnixMilli(0)
			lastRun[sample.Name] = lastRunTime
		}

		// Check if it's time to run
		if time.Since(lastRunTime) < sample.ScrapeInterval {
			return
		}

		toRun = append(toRun, func() {
			runOneScenario(sample, configFile)
			lastRun[sample.Name] = time.Now()
		})
	}

	log.Println("We have to run", len(toRun), "scenarios")

	// Create a pool of workers to run the scenarios
	pool := make(chan struct{}, maxExecutors)

	for i := 0; i < maxExecutors; i++ {
		pool <- struct{}{}
	}

	for _, run := range toRun {
		<-pool
		go func(run func()) {
			run()
			pool <- struct{}{}
		}(run)
	}
}

func metricsRecord(confPath string, maxExecutor int) {
	configFiles, err := utils.AutoDiscoverYML(confPath)
	if err != nil {
		panic(err)
	}

	log.Println("Reloading prometheus metrics")
	err = refreshPrometheusMetrics(configFiles)
	if err != nil {
		panic(err)
	}
	log.Println("Prometheus metrics reloaded")

	lastRun = make(map[string]time.Time)

	go func() {
		for {
			runScenarios(configFiles, maxExecutor)
			time.Sleep(2 * time.Second)
		}
	}()
}

func Run(wg *sync.WaitGroup, confPath string, maxExecutor, port int) {
	log.Println("Starting hidra in exporter mode")

	// Start fetching metrics
	metricsRecord(confPath, maxExecutor)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	wg.Done()
}
