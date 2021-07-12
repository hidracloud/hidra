package api

import (
	"encoding/json"
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/models"
)

func (a *API) AgentListSamples(w http.ResponseWriter, r *http.Request) {
	samples, _ := models.GetSamples()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(samples)
}
