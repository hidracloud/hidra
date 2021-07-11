package scenarios

import (
	"fmt"
	"log"
	"time"

	"github.com/JoseCarlosGarcia95/hidra/models"
)

// Run one scenario
func RunScenario(s models.Scenario) *models.ScenarioMetric {
	log.Printf("[%s] Running new scenario, \"%s\"\n", s.Name, s.Description)

	srunner := Scenarios[s.Kind]()
	srunner.Init()

	metric := models.ScenarioMetric{}
	metric.Scenario = s
	metric.StepMetrics = make([]*models.StepMetric, 0)
	metric.StartDate = time.Now()

	for _, step := range s.Steps {
		log.Printf("[%s] Running %s\n", s.Name, step.Type)
		for k, v := range step.Params {
			log.Printf("[%s]   |_ %s: %s\n", s.Name, k, v)
		}

		smetric := models.StepMetric{}
		smetric.Step = step
		smetric.StartDate = time.Now()
		err := srunner.RunStep(step.Type, step.Params)
		smetric.EndDate = time.Now()
		metric.StepMetrics = append(metric.StepMetrics, &smetric)

		if step.Negate && err == nil {
			metric.Error = fmt.Errorf("expected fail")
			metric.EndDate = time.Now()
			return &metric
		}

		if err != nil && !step.Negate {
			metric.Error = err
			metric.EndDate = time.Now()
			return &metric
		}
	}

	metric.EndDate = time.Now()
	return &metric
}

// Print scenario metrics
func PrettyPrintScenarioMetrics(m *models.ScenarioMetric) {
	log.Printf("[%s] Metrics results for: %s\n", m.Scenario.Name, m.Scenario.Description)
	if m.Error == nil {
		log.Printf("[%s] Scenario ran without issues\n", m.Scenario.Name)
	} else {
		log.Printf("[%s] Scenario ran with issues: %s\n", m.Scenario.Name, m.Error)

	}
	log.Printf("[%s] Total scenario duration: %d (ms)\n", m.Scenario.Name, m.EndDate.Sub(m.StartDate).Milliseconds())

	for _, s := range m.StepMetrics {
		log.Printf("[%s]   |_ %s duration: %d (ms)\n", m.Scenario.Name, s.Step.Type, s.EndDate.Sub(s.StartDate).Milliseconds())
	}
}
