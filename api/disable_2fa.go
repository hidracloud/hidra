package api

import (
	"net/http"

	"github.com/hidracloud/hidra/models"
)

// TwofactorQrcode Generates a QR code for the user to scan to enable 2fa.
func (a *API) DisableTwoFaConfiguration(w http.ResponseWriter, r *http.Request) {
	user := models.GetLoggedUser(r)

	err := user.Update2FAToken("")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
