package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	uuid "github.com/satori/go.uuid"
)

// Register a new agent, and generate a secret.
func (a *API) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	var registerAgentRequest RegisterAgentRequest
	var registerAgentResponse RegisterAgentResponse

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&registerAgentRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agentUuid := uuid.FromStringOrNil(params["agentid"])

	err = models.UpdateAgent(agentUuid, registerAgentRequest.Name, registerAgentRequest.Description)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	models.DeleteAgentTags(agentUuid)

	for k, v := range registerAgentRequest.Tags {
		models.CreateAgentTagByAgentID(agentUuid, k, v)
	}

	json.NewEncoder(w).Encode(registerAgentResponse)
}
