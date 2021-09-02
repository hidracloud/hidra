package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
)

// GetSampleResult is a handler to get a sample result
func (a *API) GetSampleResult(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sampleResults, err := models.GetSampleResults(params["sampleid"], 10)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sampleResults)
}
