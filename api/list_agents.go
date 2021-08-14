package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hidracloud/hidra/models"
)

// List sample response struct
type ListAgentsResponse struct {
	Id        string
	Secret    string
	Tags      map[string]string
	UpdatedAt time.Time
}

// Get a list of samples by id and checksum
func (a *API) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := models.GetAgents()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	agentResponse := make([]ListAgentsResponse, len(agents))

	for index, agent := range agents {
		tags, err := models.GetAgentTags(agent.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newAgent := ListAgentsResponse{
			Id:        agent.ID.String(),
			Tags:      make(map[string]string),
			Secret:    agent.Secret,
			UpdatedAt: agent.UpdatedAt,
		}

		for _, tag := range tags {
			newAgent.Tags[tag.Key] = tag.Value
		}

		agentResponse[index] = newAgent
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentResponse)
}
