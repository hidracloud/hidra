package api

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
)

type metricValueResponse struct {
	Value float64
	Time  int64
}

type metricResponse struct {
	MetricName string
	Labels     map[string]string
	Values     []metricValueResponse
}

type metricWithLabels struct {
	models.Metric
	Labels map[string]string
}

func calculateLabelsChecksum(labels map[string]string) string {
	checksum := ""

	keys := make([]string, 0)
	for key := range labels {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		checksum += key + ":" + labels[key]
	}
	return checksum
}

// GetMetrics returns the metrics for the given namespace and name
func (a *API) GetMetrics(w http.ResponseWriter, r *http.Request) {
	response := make([]metricResponse, 0)

	params := mux.Vars(r)

	sampleResults, err := models.GetSampleResultBySampleIDWithLimit(params["sampleid"], 10)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricsByMetricName := make(map[string][]metricWithLabels)

	for _, sampleResult := range sampleResults {
		// Get metrics by sample result
		metricsBySampleResult, err := models.GetMetricsBySampleResultID(sampleResult.ID.String())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, metric := range metricsBySampleResult {
			checksum := metric.Name
			metricLabels, err := models.GetMetricLabelByMetricID(metric.ID)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			labelsDict := make(map[string]string)

			for _, label := range metricLabels {
				labelsDict[label.Key] = label.Value
			}

			checksum += calculateLabelsChecksum(labelsDict)
			extendedMetric := metricWithLabels{
				Metric: metric,
				Labels: labelsDict,
			}

			if _, ok := metricsByMetricName[checksum]; !ok {
				metricsByMetricName[checksum] = make([]metricWithLabels, 0)
			}

			metricsByMetricName[checksum] = append(metricsByMetricName[checksum], extendedMetric)
		}
	}

	for _, metricsByName := range metricsByMetricName {
		// Latest metric labels
		latestMetric := metricsByName[0]

		metricResponse := metricResponse{
			MetricName: latestMetric.Name,
			Values:     make([]metricValueResponse, 0),
			Labels:     latestMetric.Labels,
		}

		for _, metric := range metricsByName {
			metricResponse.Values = append(metricResponse.Values, metricValueResponse{
				Value: metric.Value,
				Time:  metric.CreatedAt.Unix(),
			})
		}

		response = append(response, metricResponse)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
