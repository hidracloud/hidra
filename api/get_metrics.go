package api

import (
	"encoding/json"
	"net/http"

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

// GetMetrics returns the metrics for the given namespace and name
func (a *API) GetMetrics(w http.ResponseWriter, r *http.Request) {
	response := make([]metricResponse, 0)

	params := mux.Vars(r)
	distinctNames, err := models.GetDistinctMetricNameBySampleID(params["sampleid"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, name := range distinctNames {
		oneMetricResponse := metricResponse{}
		oneMetricResponse.MetricName = name

		checksums, err := models.GetDistinctChecksumByNameAndSampleID(name, params["sampleid"])

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, checksum := range checksums {
			metricValues, err := models.GetMetricsByNameAndSampleID(name, params["sampleid"], checksum, 10)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			labels := make(map[string]string)

			for _, metricValue := range metricValues {
				oneMetricResponse.Values = append(oneMetricResponse.Values, metricValueResponse{
					Value: metricValue.Value,
					Time:  metricValue.CreatedAt.Unix(),
				})

				metriLabels, err := models.GetMetricLabelByMetricID(metricValue.ID)

				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				for _, label := range metriLabels {
					labels[label.Key] = label.Value
				}
			}

			oneMetricResponse.Labels = labels
			response = append(response, oneMetricResponse)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
