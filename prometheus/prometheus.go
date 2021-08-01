package prometheus

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hidracloud/hidra/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	sampleMetricStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "hidra",
			Name:      "sample_metric_status",
			Help:      "This is my sample metric status",
		},
		[]string{"agent_id", "sample_id", "sample_name", "checksum"},
	)

	sampleMetricTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "hidra",
			Name:      "sample_metric_time",
			Help:      "This is my sample metric status",
		},
		[]string{"agent_id", "sample_id", "sample_name", "checksum"},
	)

	sampleStepMetricTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "hidra",
			Name:      "sample_step_metric_time",
			Help:      "This is my sample metric status",
		},
		[]string{"agent_id", "sample_id", "sample_name", "checksum", "step_name", "params", "negate"},
	)
)

func StartPrometheus(listenAddr string, pullTime int) {
	prometheus.MustRegister(sampleMetricStatus)
	prometheus.MustRegister(sampleMetricTime)
	prometheus.MustRegister(sampleStepMetricTime)

	go func() {
		for {
			samples, _ := models.GetSamples()
			agents, _ := models.GetAgents()

			for _, sample := range samples {
				sampleDef, _ := models.ReadScenariosYAML(sample.SampleData)
				for _, agent := range agents {
					lastMetricByAgent, _ := sample.GetLastMetricByAgent(agent.ID.String())

					sampleStatus := 1
					if len(lastMetricByAgent.Error) == 0 {
						sampleStatus = 0

					}

					// Write test status
					sampleMetricStatus.WithLabelValues(
						agent.ID.String(),
						sample.ID.String(),
						sample.Name,
						sample.Checksum).Set(float64(sampleStatus))

					testTime := float64(lastMetricByAgent.EndDate.UnixNano() - lastMetricByAgent.StartDate.UnixNano())
					sampleMetricTime.WithLabelValues(
						agent.ID.String(),
						sample.ID.String(),
						sample.Name,
						sample.Checksum).Set(testTime)

					stepMetrics, _ := lastMetricByAgent.GetStepMetrics()

					for i := 0; i < len(stepMetrics); i++ {
						stepMetric := stepMetrics[i]
						step := sampleDef.Scenario.Steps[i]

						paramsStr, _ := json.Marshal(step.Params)

						stepMetricTime := float64(stepMetric.EndDate.UnixNano() - stepMetric.StartDate.UnixNano())
						sampleStepMetricTime.WithLabelValues(
							agent.ID.String(),
							sample.ID.String(),
							sample.Name,
							sample.Checksum,
							step.Type,
							string(paramsStr),
							strconv.FormatBool(step.Negate)).Set(stepMetricTime)
					}
				}
			}

			time.Sleep(time.Second * time.Duration(pullTime))
		}
	}()

	// Run HTTP metrics server with auth basic
	log.Println("Starting metrics at " + listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
