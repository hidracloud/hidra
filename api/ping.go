package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/utils"
)

// Response for ping req
type PingResponse struct {
	Pong    bool
	Version string
}

// This method return Pong
func (a *API) Ping(w http.ResponseWriter, r *http.Request) {
	pingResponse := PingResponse{Pong: true, Version: utils.Version()}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingResponse)
}
