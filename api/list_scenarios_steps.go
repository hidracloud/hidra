package api

import (
	"encoding/json"
	"net/http"

	"github.com/hidracloud/hidra/models"
	"github.com/hidracloud/hidra/scenarios"
)

type ScenarioResponse struct {
	Description     string
	StepDefinitions map[string]models.StepDefinition
}

// This method scenario and step lists.
func (a *API) ListScenariosSteps(w http.ResponseWriter, r *http.Request) {
	scenarios := scenarios.GetAll()

	scenarioResponse := make(map[string]ScenarioResponse)

	for name, scenario := range scenarios {
		oneScenario := scenario()
		oneScenario.Init()

		scenarioResponse[name] = ScenarioResponse{
			Description:     oneScenario.Description(),
			StepDefinitions: oneScenario.GetScenarioDefinitions(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scenarioResponse)
}
