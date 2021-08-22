package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
)

type MetricValueResponse struct {
	Value float64
	Time  int64
}

type MetricResponse struct {
	MetricName string
	Labels     map[string]string
	Values     []MetricValueResponse
}

func (a *API) GetMetrics(w http.ResponseWriter, r *http.Request) {
	response := make([]MetricResponse, 0)

	params := mux.Vars(r)
	distinctNames, err := models.GetDistinctMetricNameBySampleId(params["sampleid"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, name := range distinctNames {
		oneMetricResponse := MetricResponse{}
		oneMetricResponse.MetricName = name

		metricValues, err := models.GetMetricsByNameAndSampleID(name, params["sampleid"], 10)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		labels := make(map[string]string)

		for _, metricValue := range metricValues {
			oneMetricResponse.Values = append(oneMetricResponse.Values, MetricValueResponse{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
