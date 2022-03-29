package scenarios

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hidracloud/hidra/pkg/models"
	"github.com/hidracloud/hidra/pkg/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	msgTotalScenarioDurationFmt = "[%s] Total scenario duration: %d (ms)\n"
)

// InitializeScenario initialize a new scenario
func InitializeScenario(s models.Scenario) (models.IScenario, error) {
	if len(s.Kind) == 0 {
		return nil, fmt.Errorf("Scenario kind \"%s\" is not supported", s.Kind)
	}

	if _, ok := Sample[s.Kind]; !ok {
		return nil, fmt.Errorf("Scenario kind \"%s\" is not supported", s.Kind)
	}

	srunner := Sample[s.Kind]()
	srunner.Init()

	return srunner, nil
}

// RunIScenario run already initialize scenario
func RunIScenario(ctx context.Context, name, desc string, s models.Scenario, srunner models.IScenario) *models.ScenarioResult {
	metric := models.ScenarioResult{}
	metric.Scenario = s
	metric.StepResults = make([]*models.StepResult, 0)
	metric.StartDate = time.Now()

	defer srunner.Close()

	ctx, span := otel.Tracer("scenarios").Start(ctx, name, trace.WithAttributes(
		attribute.String("name", name),
		attribute.String("description", desc),
		attribute.String("kind", s.Kind),
	))
	defer span.End()

	for _, step := range s.Steps {
		smetric := models.StepResult{}
		smetric.Step = step
		smetric.StartDate = time.Now()
		customMetrics, err := srunner.RunStep(ctx, step.Type, step.Params, step.Timeout)

		smetric.Metrics = customMetrics
		smetric.EndDate = time.Now()
		metric.StepResults = append(metric.StepResults, &smetric)

		if step.Negate && err == nil {
			metric.Error = fmt.Errorf("expected fail")
			metric.EndDate = time.Now()
			AddScreenshotsToQueue(&metric, s, name, desc)
			span.SetStatus(codes.Error, metric.Error.Error())
			return &metric
		}

		if err != nil && !step.Negate {
			metric.Error = err
			metric.EndDate = time.Now()
			AddScreenshotsToQueue(&metric, s, name, desc)
			span.SetStatus(codes.Error, metric.Error.Error())
			return &metric
		}
	}

	metric.EndDate = time.Now()
	return &metric
}

// RunScenario Run one scenario
func RunScenario(ctx context.Context, s models.Scenario, name, desc string) (*models.ScenarioResult, error) {
	utils.LogDebug("[%s] Running new scenario, \"%s\"\n", name, desc)
	srunner, err := InitializeScenario(s)
	if err != nil {
		return nil, err
	}
	return RunIScenario(ctx, name, desc, s, srunner), nil
}

// PrettyPrintScenarioResults Print scenario metrics
func PrettyPrintScenarioResults(m *models.ScenarioResult, name, desc string) {
	log.Printf("[%s] Metrics results for: %s\n", name, desc)
	if m.Error == nil {
		log.Printf("[%s] Scenario ran without issues\n", name)
	} else {
		log.Printf("[%s] Scenario ran with issues: %s\n", name, m.Error)
	}

	log.Printf(msgTotalScenarioDurationFmt, name, m.EndDate.Sub(m.StartDate).Milliseconds())

	for _, s := range m.StepResults {
		log.Printf("[%s]   |_ %s duration: %d (ms)\n", name, s.Step.Type, s.EndDate.Sub(s.StartDate).Milliseconds())
		// Print metrics
		for _, m := range s.Metrics {
			log.Printf("[%s]     |_ %s: (%v) %f\n", name, m.Name, m.Labels, m.Value)
		}

	}
}

// PrettyPrintScenarioResults2String Print scenario metrics
func PrettyPrintScenarioResults2String(m *models.ScenarioResult, name, desc string) string {
	out := ""

	out += fmt.Sprintf("[%s] Metrics results for: %s\n", name, desc)
	if m.Error == nil {
		out += fmt.Sprintf("[%s] Scenario ran without issues\n", name)
	} else {
		out += fmt.Sprintf("[%s] Scenario ran with issues: %s\n", name, m.Error)

	}

	out += fmt.Sprintf(msgTotalScenarioDurationFmt, name, m.EndDate.Sub(m.StartDate).Milliseconds())

	for _, s := range m.StepResults {
		out += fmt.Sprintf("[%s]   |_ %s duration: %d (ms)\n", name, s.Step.Type, s.EndDate.Sub(s.StartDate).Milliseconds())
	}

	out += fmt.Sprintf(msgTotalScenarioDurationFmt, name, m.EndDate.Sub(m.StartDate).Milliseconds())
	return out
}
