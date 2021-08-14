package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	uuid "github.com/satori/go.uuid"
)

// Get a list of samples by id and checksum
func (a *API) GetAgent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	agent, err := models.GetAgent(uuid.FromStringOrNil(params["agentid"]))
	if err != nil || agent.ID == uuid.Nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agentResponse := ListAgentsResponse{}

	tags, err := models.GetAgentTags(agent.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newAgent := ListAgentsResponse{
		Id:     agent.ID.String(),
		Tags:   make(map[string]string),
		Secret: agent.Secret,
	}

	for _, tag := range tags {
		newAgent.Tags[tag.Key] = tag.Value
	}

	agentResponse = newAgent

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentResponse)
}
