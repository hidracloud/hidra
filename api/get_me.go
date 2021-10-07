package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
)

// GetMe return user info of logged in user
func (api *API) GetMe(w http.ResponseWriter, r *http.Request) {
	user := models.GetLoggedUser(r)

	w.Header().Set("Content-Type", "application/json")

	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}
