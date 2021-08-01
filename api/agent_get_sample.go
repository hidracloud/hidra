package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
)

// Get one sample from agent
func (a *API) AgentGetSample(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sample, _ := models.GetSampleById(params["sampleid"])
	w.Write(sample.SampleData)
}
