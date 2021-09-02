package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
)

// AgentGetSample is a handler that returns a sample of the agent
func (a *API) AgentGetSample(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sample, _ := models.GetSampleByID(params["sampleid"])
	w.Write(sample.SampleData)
}
