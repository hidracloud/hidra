package api

import (
	"encoding/json"
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/models"
)

// Response for agent token req
type AgentTokenResponse struct {
	AgentToken string
}

// Generate a temporal agent token
func (a *API) AgentToken(w http.ResponseWriter, r *http.Request) {
	user := models.GetLoggedUser(r)

	err := models.CheckIfAllowTo(user, "new_agent_token")

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	agentToken, err := models.CreateRegisterAgentToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	agentTokenResponse := AgentTokenResponse{AgentToken: agentToken}
	json.NewEncoder(w).Encode(agentTokenResponse)
}
