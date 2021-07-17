package api

import (
	"encoding/json"
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/gorilla/mux"
)

// Recieve new metrics from an agent
func (a *API) AgentPushMetrics(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var scenarioMetric models.ScenarioMetric
	sample, _ := models.GetSampleById(params["sampleid"])

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&scenarioMetric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sample.PushMetrics(&scenarioMetric)

}
