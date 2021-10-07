package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"golang.org/x/crypto/bcrypt"
)

type updatePasswordRequest struct {
	OldPassword string
	NewPassword string
}

// UpdatePassword update password
func (api *API) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var updatePasswordRequest updatePasswordRequest

	user := models.GetLoggedUser(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&updatePasswordRequest)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if updatePasswordRequest.OldPassword == "" || updatePasswordRequest.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(updatePasswordRequest.OldPassword))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = user.UpdatePassword(updatePasswordRequest.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
