package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hidracloud/hidra/models"
)

// RegisterSample is a function to register a sample
func (a *API) RegisterSample(w http.ResponseWriter, r *http.Request) {
	user := models.GetLoggedUser(r)

	err := models.CheckIfAllowTo(user, "register_sample")

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	sample, err := models.ReadScenariosYAML(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	newSample, err := models.RegisterSample(sample.Name, sample.Description, sample.Scenario.Kind, data, user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newSample)
}
