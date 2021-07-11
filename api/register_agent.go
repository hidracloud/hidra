package api

import (
	"encoding/json"
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/utils"
)

type RegisterAgentRequest struct {
	Host string
	Port uint
	Tags map[string]string
}

type RegisterAgentResponse struct {
	Secret string
}

func (a *API) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var registerAgentRequest RegisterAgentRequest
	var registerAgentResponse RegisterAgentResponse

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&registerAgentRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	registerAgentResponse.Secret = utils.RandString(128)
	json.NewEncoder(w).Encode(registerAgentResponse)
}
