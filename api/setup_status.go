package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
)

// SetupStatusResponse is the response of the setup status
type SetupStatusResponse struct {
	Status bool
}

// SetupStatus returns the status of the setup
func (a *API) SetupStatus(w http.ResponseWriter, r *http.Request) {
	usersCount := models.GetUserCount()

	setupStatus := SetupStatusResponse{Status: usersCount != 0}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setupStatus)
}
