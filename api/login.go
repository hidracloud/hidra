package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/hidracloud/hidra/models"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AuthToken string `json:"AuthToken,omitempty"`
	Error     string `json:"Error,omitempty"`
}

// Generate a login session
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	var loginResponse LoginResponse

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		loginResponse.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse)

		return
	}

	n := rand.Intn(1000)
	time.Sleep(time.Duration(n) * time.Millisecond)

	user := models.GetUserByEmail(loginRequest.Email)

	if user.ID == uuid.Nil {
		loginResponse.Error = "BAD_CREDENTIALS"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse)

		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(loginRequest.Password))

	if err != nil {
		loginResponse.Error = "BAD_CREDENTIALS"
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse)

		return
	}

	token, err := models.CreateUserToken(user)

	if err != nil {
		loginResponse.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(loginResponse)

		return
	}

	loginResponse.AuthToken = token
	json.NewEncoder(w).Encode(loginResponse)
}
