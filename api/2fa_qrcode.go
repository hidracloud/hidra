package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"github.com/xlzd/gotp"
)

type twoFactorResponse struct {
	URI       string
	MFASecret string
}

// TwofactorQrcode Generates a QR code for the user to scan to enable 2fa.
func (a *API) TwofactorQrcode(w http.ResponseWriter, r *http.Request) {
	user := models.GetLoggedUser(r)

	if user == nil || user.TwoFactorToken != "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	mfaSecret := gotp.RandomSecret(16)

	totp := gotp.NewDefaultTOTP(mfaSecret)
	totp.Now()

	twoFactorResponse := twoFactorResponse{
		URI:       totp.ProvisioningUri(user.Email, "Hidra"),
		MFASecret: mfaSecret,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(twoFactorResponse)
}
