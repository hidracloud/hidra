package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/utils"
)

// Response for ping req
type SystemInfoResponse struct {
	Version string
	DbType  string
}

// This method return Pong
func (a *API) SystemInfo(w http.ResponseWriter, r *http.Request) {
	systemInfoResponse := SystemInfoResponse{}

	// Retrieve hidra version
	systemInfoResponse.Version = utils.Version()
	systemInfoResponse.DbType = a.dbType
	// Get database type

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemInfoResponse)
}
