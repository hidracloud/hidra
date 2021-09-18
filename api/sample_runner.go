package api

import (
	"io/ioutil"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
)

func (a *API) SampleRunner(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	slist, err := models.ReadScenariosYAML(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m := scenarios.RunScenario(slist.Scenario, slist.Name, slist.Description)
	w.Write([]byte(scenarios.PrettyPrintScenarioResults2String(m, slist.Name, slist.Description)))
}
