package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
)

type scenarioResponse struct {
	Description     string
	StepDefinitions map[string]models.StepDefinition
}

// ListScenariosSteps returns a list of all the steps of all the scenarios
func (a *API) ListScenariosSteps(w http.ResponseWriter, r *http.Request) {
	scenarios := scenarios.GetAll()

	allScenarioResponse := make(map[string]scenarioResponse)

	for name, scenario := range scenarios {
		oneScenario := scenario()
		oneScenario.Init()

		allScenarioResponse[name] = scenarioResponse{
			Description:     oneScenario.Description(),
			StepDefinitions: oneScenario.GetScenarioDefinitions(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allScenarioResponse)
}
