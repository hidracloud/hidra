package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hidracloud/hidra/models"
	uuid "github.com/satori/go.uuid"
)

// UpdateSample is a function to update a sample
func (a *API) UpdateSample(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

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

	oldSample, err := models.GetSampleByID(params["sampleid"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if oldSample.ID == uuid.Nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	sample, err := models.ReadScenariosYAML(data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	updateSample, err := models.UpdateSample(params["sampleid"], sample.Name, sample.Description, sample.Scenario.Kind, data, user)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updateSample)
}
