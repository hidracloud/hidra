package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/hidracloud/hidra/models"
	"github.com/xlzd/gotp"
)

type twoFaConfiguration struct {
	CurrentToken string
	MFASecret    string
}

// TwofactorQrcode Generates a QR code for the user to scan to enable 2fa.
func (a *API) TwoFaConfiguration(w http.ResponseWriter, r *http.Request) {
	var twoFaConfiguration twoFaConfiguration

	user := models.GetLoggedUser(r)

	if user == nil || user.TwoFactorToken != "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&twoFaConfiguration)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	totp := gotp.NewDefaultTOTP(twoFaConfiguration.MFASecret)
	totp.Now()

	if totp.Verify(twoFaConfiguration.CurrentToken, int(time.Now().Unix())) {
		err = user.Update2FAToken(twoFaConfiguration.MFASecret)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
