package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
)

// GetAgent is a function to get agent
func (a *API) DeleteSample(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	user := models.GetLoggedUser(r)

	err := models.CheckIfAllowTo(user, "delete_sample")

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = models.DeleteMetricBySampleID(params["sample_id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = models.DeleteAllSampleResults(params["sample_id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = models.DeleteSample(params["sampleid"])
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
