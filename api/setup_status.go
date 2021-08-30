package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
)

// Response for ping req
type SetupStatusResponse struct {
	Status bool
}

func (a *API) SetupStatus(w http.ResponseWriter, r *http.Request) {
	usersCount := models.GetUserCount()

	setupStatus := SetupStatusResponse{Status: usersCount != 0}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setupStatus)
}
