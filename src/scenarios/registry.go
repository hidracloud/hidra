// Package scenarios contains all scenarios
package scenarios

import "github.com/hidracloud/hidra/src/models"

// ScenarioGenerator interface
type ScenarioGenerator func() models.IScenario

// Scenarios contains all scenarios
var Scenarios = map[string]ScenarioGenerator{}

// Add new scenario
func Add(name string, scenario ScenarioGenerator) {
	Scenarios[name] = scenario
}

// GetAll scenarios
func GetAll() map[string]ScenarioGenerator {
	return Scenarios
}
