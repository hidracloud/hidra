package api

import (
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/models"
	"github.com/gorilla/mux"
)

// Get one sample from agent
func (a *API) AgentGetSample(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sample, _ := models.GetSampleById(params["sampleid"])
	w.Write(sample.SampleData)
}
