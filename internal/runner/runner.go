package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/misc"
	"github.com/hidracloud/hidra/v3/internal/plugins"
	"github.com/hidracloud/hidra/v3/internal/report"
	"github.com/hidracloud/hidra/v3/internal/utils"

	log "github.com/sirupsen/logrus"
)

// StepParamTemplate represents a step parameter template.
type StepParamTemplate struct {
	Env       map[string]string
	Date      time.Time
	Context   context.Context
	Variables map[string]string
}

// RunnerResult represents the result of a runner.
type RunnerResult struct {
	Context context.Context
	Metrics []*metrics.Metric
	Reports []*report.Report
	Error   error
}

// GetContext return context value by key
func (s *StepParamTemplate) GetContext(key string) string {
	if s.Context == nil {
		return ""
	}

	value := s.Context.Value(key)

	if value == nil {
		return ""
	}

	return value.(string)
}

// Replace replaces the template.
func (s *StepParamTemplate) Replace(m map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range m {
		t, err := template.New("").Parse(v)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, s)
		if err != nil {
			return nil, err
		}

		result[k] = buf.String()
	}
	return result, nil
}

// RunWithVariables runs the step with variables.
func RunWithVariables(ctx context.Context, variables map[string]string, sample *config.SampleConfig) (context.Context, []*metrics.Metric, *report.Report, error) {
	var allMetrics, newMetrics []*metrics.Metric

	depthSize := 1
	lastPlugin := ""
	pluginsByNames := make(map[string]plugins.PluginInterface)

	stepParamTemplate := StepParamTemplate{
		Env:     utils.EnvToMap(),
		Date:    time.Now(),
		Context: ctx,
	}

	variables, err := stepParamTemplate.Replace(variables)

	if err != nil {
		return ctx, nil, nil, err
	}

	stepParamTemplate.Variables = variables

	// cleanup
	defer func() {
		for _, plugin := range pluginsByNames {
			if plugin.StepExists("onClose") {
				ctx, _, err = plugin.RunStep(ctx, &plugins.Step{
					Name: "onClose",
					Args: map[string]string{},
				})

				if err != nil {
					log.Warnf("Error closing plugin: %v", err)
					err = nil
				}
			}
		}

		misc.CleanupContext(ctx)
	}()

	startTime := time.Now()
	stepCounter := 0
	for _, step := range sample.Steps {
		// Check if timeout is reached in context, if so, stop the execution
		if ctx.Err() != nil {
			log.Warnf("Timeout reached, stopping execution of sample %s", sample.Name)
			return ctx, allMetrics, nil, ctx.Err()
		}

		log.Debugf("|%s Running plugin %s", strings.Repeat("_", depthSize), step.Plugin)
		log.Debugf("|_%s Action: %v", strings.Repeat("_", depthSize), step.Action)
		log.Debugf("|_%s Parameters: ", strings.Repeat("_", depthSize))

		stepParamTemplate.Context = ctx

		step.Parameters, err = stepParamTemplate.Replace(step.Parameters)

		if err != nil {
			return ctx, nil, nil, err
		}

		for k, v := range step.Parameters {
			log.Debugf("|__%s %s: %v", strings.Repeat("_", depthSize), k, v)
		}

		depthSize++
		if step.Plugin == "" {
			step.Plugin = lastPlugin
		}
		plugin := plugins.GetPlugin(step.Plugin)

		if plugin == nil {
			return ctx, allMetrics, nil, fmt.Errorf("plugin %s not found", step.Plugin)
		}

		lastPlugin = step.Plugin
		pluginsByNames[step.Plugin] = plugin

		ctx, newMetrics, err = plugin.RunStep(ctx, &plugins.Step{
			Name:   step.Action,
			Args:   step.Parameters,
			Negate: step.Negate,
		})

		allMetrics = append(allMetrics, newMetrics...)

		if err != nil {
			err = fmt.Errorf("%s#%d: %s", sample.Path, stepCounter, err)
			return ctx, allMetrics, report.NewReport(sample, allMetrics, variables, time.Since(startTime), ctx, err), err
		}

		stepCounter++
	}

	return ctx, allMetrics, nil, err
}

// RunSample runs a sample.
func RunSample(ctx context.Context, sample *config.SampleConfig) *RunnerResult {
	var err error

	allMetrics := []*metrics.Metric{}
	allReports := []*report.Report{}

	ctx = context.WithValue(ctx, misc.ContextTimeout, sample.Timeout)

	for _, variables := range sample.Variables {
		ctx, newMetrics, report, err := RunWithVariables(ctx, variables, sample)

		allMetrics = append(allMetrics, newMetrics...)
		allReports = append(allReports, report)
		if err != nil {
			return &RunnerResult{
				Context: ctx,
				Metrics: allMetrics,
				Reports: allReports,
				Error:   err,
			}
		}

	}
	return &RunnerResult{
		Context: ctx,
		Metrics: allMetrics,
		Reports: allReports,
		Error:   err,
	}
}
