// Scenario base package for autoregistering
package scenarios

import "github.com/hidracloud/hidra/models"

type ScenarioGenerator func() models.IScenario

var Scenarios = map[string]ScenarioGenerator{}

// Add new scenario
func Add(name string, scenario ScenarioGenerator) {
	Scenarios[name] = scenario
}
