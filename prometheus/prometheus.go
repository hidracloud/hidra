package prometheus

import (
	"log"
	"net/http"
	"time"

	"github.com/hidracloud/hidra/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartPrometheus starts the prometheus server
func StartPrometheus(listenAddr string, pullTime int) {
	go func() {
		gaugeDict := make(map[string]*prometheus.GaugeVec)
		labelSet := make(map[string][]string, 0)

		for {
			metricsNames, err := models.GetDistinctMetricName()

			if err != nil {
				log.Println(err)
				continue
			}

			// Search for new metrics
			for _, metricName := range metricsNames {
				if _, ok := gaugeDict[metricName]; !ok {
					oneMetric, err := models.GetMetricByName(metricName)
					if err != nil {
						log.Println(err)
						continue
					}

					metriLabels, err := models.GetMetricLabelByMetricID(oneMetric.ID)
					if err != nil {
						log.Println(err)
						continue
					}

					labels := []string{}

					for _, label := range metriLabels {
						labels = append(labels, label.Key)
					}

					helpText := oneMetric.Description

					if helpText == "" {
						helpText = "Auto generated metric by hidra"
					}

					gaugeDict[metricName] = prometheus.NewGaugeVec(
						prometheus.GaugeOpts{
							Namespace: "hidra",
							Name:      metricName,
							Help:      helpText,
						},
						labels,
					)

					labelSet[metricName] = labels

					prometheus.MustRegister(gaugeDict[metricName])
				}

				gaugeDict[metricName].Reset()

				distinctLabels, err := models.GetDistinctChecksumByName(metricName)
				if err != nil {
					log.Println(err)
					continue
				}

				for _, label := range distinctLabels {
					oneMetric, err := models.GetMetricByChecksum(label, metricName)
					if err != nil {
						log.Println(err)
						continue
					}

					labelsDict := make(map[string]string)

					metricLabels, err := models.GetMetricLabelByMetricID(oneMetric.ID)
					if err != nil {
						log.Println(err)
						continue
					}

					for _, label := range metricLabels {
						labelsDict[label.Key] = label.Value
					}

					labels := make([]string, len(labelSet[metricName]))
					for k, v := range labelSet[metricName] {
						if _, ok := labelsDict[v]; ok {
							labels[k] = labelsDict[v]
						}
					}

					gaugeDict[metricName].WithLabelValues(labels...).Set(oneMetric.Value)
				}
			}

			time.Sleep(time.Second * time.Duration(pullTime))
		}
	}()

	log.Println("Starting metrics at " + listenAddr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
