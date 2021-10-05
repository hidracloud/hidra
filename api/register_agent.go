package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/utils"
)

// RegisterAgentRequest is the request to register a new agent
type RegisterAgentRequest struct {
	Name        string
	Description string
	Tags        map[string]string
}

// RegisterAgentResponse is the response of RegisterAgent
type RegisterAgentResponse struct {
	Secret string
}

// RegisterAgent register a new agent
func (a *API) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var registerAgentRequest RegisterAgentRequest
	var registerAgentResponse RegisterAgentResponse

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&registerAgentRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agentSecret := utils.RandString(32)

	err = models.CreateAgent(agentSecret, registerAgentRequest.Name, registerAgentRequest.Description, registerAgentRequest.Tags)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	registerAgentResponse.Secret = agentSecret

	json.NewEncoder(w).Encode(registerAgentResponse)
}
