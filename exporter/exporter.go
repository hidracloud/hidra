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

var hidraCustomMetrics map[string]*prometheus.GaugeVec

func refreshPrometheusMetrics(configFiles []string) error {
	// hidraScenarioStatusVec := prometheus.NewGaugeVec()
	prometheusLabels = []string{"name", "description", "kind"}
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
	hidraCustomMetrics = make(map[string]*prometheus.GaugeVec)

	// Restart prometheus
	prometheus.MustRegister(hidraScenarioStatusVec)
	prometheus.MustRegister(hidraScenarioElapsedVec)
	prometheus.MustRegister(hidraStepStatusVec)
	prometheus.MustRegister(hidraStepElapsedVec)

	return nil
}

func readLabels(sample *models.Scenarios) []string {
	labels := []string{}

	labels = append(labels, sample.Name)
	labels = append(labels, sample.Description)
	labels = append(labels, sample.Scenario.Kind)

	for _, label := range prometheusLabels[3:] {
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

func runScenarios(configFiles []string) {
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
			continue
		}

		m := scenarios.RunScenario(sample.Scenario, sample.Name, sample.Description)
		lastRun[sample.Name] = time.Now()

		status := 0

		if m.Error == nil {
			status = 1
		}

		labels := readLabels(sample)
		hidraScenarioStatusVec.WithLabelValues(labels...).Set(float64(status))
		hidraScenarioElapsedVec.WithLabelValues(labels...).Set(float64(m.EndDate.UnixMilli() - m.StartDate.UnixMilli()))

		for _, step := range m.StepResults {
			stepLabels := append(labels, step.Step.Type)

			for _, metric := range step.Metrics {
				// Check if custom metric exists
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

				metricLabels := []string{}
				metricLabels = append(metricLabels, stepLabels...)

				for _, label := range metric.Labels {
					metricLabels = append(metricLabels, label)
				}

				hidraCustomMetrics[metric.Name].WithLabelValues(metricLabels...).Set(float64(metric.Value))
			}

			hidraStepElapsedVec.WithLabelValues(stepLabels...).Set(float64(step.EndDate.UnixMilli() - step.StartDate.UnixMilli()))
		}
	}
}

func metricsRecord(confPath string) {
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
			runScenarios(configFiles)
			time.Sleep(2 * time.Second)
		}
	}()
}

func Run(wg *sync.WaitGroup, confPath string) {
	log.Println("Starting hidra in exporter mode")

	// Start fetching metrics
	metricsRecord(confPath)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
	wg.Done()
}
