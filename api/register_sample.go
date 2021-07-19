package api

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/JoseCarlosGarcia95/hidra/models"
)

// Generate a new sample
func (a *API) RegisterSample(w http.ResponseWriter, r *http.Request) {
	user := models.GetLoggedUser(r)

	err := models.CheckIfAllowTo(user, "register_sample")

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	data, err := ioutil.ReadAll(r.Body)

	sample, err := models.ReadScenariosYAML(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = models.RegisterSample(sample.Name, data, user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
}
