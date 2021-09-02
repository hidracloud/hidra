package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/utils"
)

type systemInfoResponse struct {
	Version string
	DbType  string
}

// SystemInfo is the response of the system info
func (a *API) SystemInfo(w http.ResponseWriter, r *http.Request) {
	systemInfoResponse := systemInfoResponse{}

	// Retrieve hidra version
	systemInfoResponse.Version = utils.Version()
	systemInfoResponse.DbType = a.dbType

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(systemInfoResponse)
}
