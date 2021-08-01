package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/utils"
)

type RegisterAgentRequest struct {
	Tags map[string]string
}

type RegisterAgentResponse struct {
	Secret string
}

// Register a new agent, and generate a secret.
func (a *API) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var registerAgentRequest RegisterAgentRequest
	var registerAgentResponse RegisterAgentResponse

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&registerAgentRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agentSecret := utils.RandString(256)

	err = models.CreateAgent(agentSecret, registerAgentRequest.Tags)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	registerAgentResponse.Secret = agentSecret

	json.NewEncoder(w).Encode(registerAgentResponse)
}
