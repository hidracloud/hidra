package api

import (
	"encoding/json"
	"net/http"
)

// Response for ping req
type pingResponse struct {
	Pong bool
}

// Ping is a simple ping endpoint
func (a *API) Ping(w http.ResponseWriter, r *http.Request) {
	pingResponse := pingResponse{Pong: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingResponse)
}
