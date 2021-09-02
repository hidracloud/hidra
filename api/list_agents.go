package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hidracloud/hidra/models"
)

// ListAgentsResponse is the response of the list agents
type ListAgentsResponse struct {
	ID          string
	Name        string
	Description string
	Secret      string
	Tags        map[string]string
	UpdatedAt   time.Time
}

// ListAgents returns the list of agents
func (a *API) ListAgents(w http.ResponseWriter, r *http.Request) {
	var err error
	var agents []models.Agent

	search := r.URL.Query().Get("search")

	if search == "" {
		agents, err = models.GetAgents()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		agents, err = models.SearchAgentByName(search)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	agentResponse := make([]ListAgentsResponse, len(agents))

	for index, agent := range agents {
		tags, err := models.GetAgentTags(agent.ID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newAgent := ListAgentsResponse{
			ID:          agent.ID.String(),
			Name:        agent.Name,
			Description: agent.Description,
			Tags:        make(map[string]string),
			Secret:      agent.Secret,
			UpdatedAt:   agent.UpdatedAt,
		}

		for _, tag := range tags {
			newAgent.Tags[tag.Key] = tag.Value
		}

		agentResponse[index] = newAgent
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agentResponse)
}
