// Package scenarios contains all scenarios
package scenarios

import "github.com/hidracloud/hidra/pkg/models"

// ScenarioGenerator interface
type ScenarioGenerator func() models.IScenario

// Sample contains all scenarios
var Sample = map[string]ScenarioGenerator{}

// Add new scenario
func Add(name string, scenario ScenarioGenerator) {
	Sample[name] = scenario
}

// GetAll scenarios
func GetAll() map[string]ScenarioGenerator {
	return Sample
}
