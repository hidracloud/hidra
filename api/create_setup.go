package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
)

type createSetupRequest struct {
	Email    string
	Password string
}

// CreateSetup is the handler to create a new setup
func (a *API) CreateSetup(w http.ResponseWriter, r *http.Request) {
	usersCount := models.GetUserCount()

	if usersCount != 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req createSetupRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := models.CreateUser(req.Email, req.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	models.AddPermission2User(user, "superadmin")

	setupStatus := SetupStatusResponse{Status: true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setupStatus)
}
