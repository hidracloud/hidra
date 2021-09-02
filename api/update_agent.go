package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	uuid "github.com/satori/go.uuid"
)

// UpdateAgent update agent
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

	agentUUID := uuid.FromStringOrNil(params["agentid"])

	err = models.UpdateAgent(agentUUID, registerAgentRequest.Name, registerAgentRequest.Description)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	models.DeleteAgentTags(agentUUID)

	for k, v := range registerAgentRequest.Tags {
		models.CreateAgentTagByAgentID(agentUUID, k, v)
	}

	json.NewEncoder(w).Encode(registerAgentResponse)
}
