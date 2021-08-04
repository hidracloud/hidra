package scenarios

import (
	"fmt"
	"log"
	"time"

	"github.com/hidracloud/hidra/models"
)

// Run one scenario
func RunScenario(s models.Scenario, name, desc string) *models.ScenarioMetric {
	log.Printf("[%s] Running new scenario, \"%s\"\n", name, desc)

	srunner := Scenarios[s.Kind]()
	srunner.Init()

	metric := models.ScenarioMetric{}
	metric.Scenario = s
	metric.StepMetrics = make([]*models.StepMetric, 0)
	metric.StartDate = time.Now()

	for _, step := range s.Steps {
		smetric := models.StepMetric{}
		smetric.Step = step
		smetric.StartDate = time.Now()
		custom_metrics, err := srunner.RunStep(step.Type, step.Params)

		for _, m := range custom_metrics {
			log.Printf("Custom metric `%s` with value %f & labels %s\n", m.Name, m.Value, m.Labels)
		}

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
func PrettyPrintScenarioMetrics(m *models.ScenarioMetric, name, desc string) {
	log.Printf("[%s] Metrics results for: %s\n", name, desc)
	if m.Error == nil {
		log.Printf("[%s] Scenario ran without issues\n", name)
	} else {
		log.Printf("[%s] Scenario ran with issues: %s\n", name, m.Error)

	}
	log.Printf("[%s] Total scenario duration: %d (ms)\n", name, m.EndDate.Sub(m.StartDate).Milliseconds())

	for _, s := range m.StepMetrics {
		log.Printf("[%s]   |_ %s duration: %d (ms)\n", name, s.Step.Type, s.EndDate.Sub(s.StartDate).Milliseconds())
	}
}
