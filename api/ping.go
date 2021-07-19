package api

import (
	"encoding/json"
	"net/http"
)

// Response for ping req
type PingResponse struct {
	Pong bool
}

// This method return Pong
func (a *API) Ping(w http.ResponseWriter, r *http.Request) {
	pingResponse := PingResponse{Pong: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingResponse)
}
