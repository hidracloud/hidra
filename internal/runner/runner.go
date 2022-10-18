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
	Metrics []*metrics.Metric
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

// RestoreOriginParamsMetrics replaces the parameters in the metrics.
func RestoreOriginParamsMetrics(metrics []*metrics.Metric, params map[string]string) []*metrics.Metric {
	for _, metric := range metrics {
		for k, v := range params {
			if _, ok := metric.Labels[k]; ok {
				metric.Labels[k] = v
			}
		}
	}
	return metrics
}

// RunWithVariables runs the step with variables.
func RunWithVariables(ctx context.Context, variables map[string]string, stepsgen map[string]any, sample *config.SampleConfig) ([]*metrics.Metric, error) {
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
		return nil, err
	}

	stepParamTemplate.Variables = variables

	// cleanup
	defer func() {
		for _, plugin := range pluginsByNames {
			if plugin.StepExists("onClose") {
				_, err = plugin.RunStep(ctx, stepsgen, &plugins.Step{
					Name: "onClose",
					Args: map[string]string{},
				})

				if err != nil {
					log.Warnf("Error closing plugin: %v", err)
					err = nil
				}
			}
		}

		for step := range stepsgen {
			switch step {
			case misc.ContextAttachment:
				attachments := stepsgen[step].(map[string][]byte)
				for name := range attachments {
					delete(attachments, name)
				}
				delete(stepsgen, step)
			default:
				delete(stepsgen, step)
			}
		}
	}()

	startTime := time.Now()
	stepCounter := 0
	for _, step := range sample.Steps {
		// Check if timeout is reached in context, if so, stop the execution
		if ctx.Err() != nil {
			log.Warnf("Timeout reached, stopping execution of sample %s", sample.Name)
			return allMetrics, ctx.Err()
		}

		log.Debugf("|%s Running plugin %s", strings.Repeat("_", depthSize), step.Plugin)
		log.Debugf("|_%s Action: %v", strings.Repeat("_", depthSize), step.Action)
		log.Debugf("|_%s Parameters: ", strings.Repeat("_", depthSize))

		stepParamTemplate.Context = ctx

		params, err := stepParamTemplate.Replace(step.Parameters)

		if err != nil {
			return nil, err
		}

		for k, v := range params {
			log.Debugf("|__%s %s: %v", strings.Repeat("_", depthSize), k, v)
		}

		depthSize++
		if step.Plugin == "" {
			step.Plugin = lastPlugin
		}
		plugin := plugins.GetPlugin(step.Plugin)

		if plugin == nil {
			return allMetrics, fmt.Errorf("plugin %s not found", step.Plugin)
		}

		lastPlugin = step.Plugin
		pluginsByNames[step.Plugin] = plugin

		newMetrics, err = plugin.RunStep(ctx, stepsgen, &plugins.Step{
			Name:   step.Action,
			Args:   params,
			Negate: step.Negate,
		})

		originMetrics := RestoreOriginParamsMetrics(newMetrics, step.Parameters)
		allMetrics = append(allMetrics, originMetrics...)

		if err != nil {
			err = fmt.Errorf("%s#%d: %s", sample.Path, stepCounter, err)
			report := report.NewReport(sample, allMetrics, variables, time.Since(startTime), stepsgen, err)
			rErr := report.Save()
			if rErr != nil {
				log.Warn(rErr)
			}
			return allMetrics, err
		}

		stepCounter++
	}

	return allMetrics, err
}

// RunSample runs a sample.
func RunSample(ctx context.Context, sample *config.SampleConfig) *RunnerResult {
	var err error

	allMetrics := []*metrics.Metric{}

	stepsgen := make(map[string]any, 0)
	stepsgen[misc.ContextAttachment] = make(map[string][]byte)
	stepsgen[misc.ContextTimeout] = sample.Timeout

	for _, variables := range sample.Variables {
		newMetrics, err := RunWithVariables(ctx, variables, stepsgen, sample)
		allMetrics = append(allMetrics, newMetrics...)

		if err != nil {
			return &RunnerResult{
				Metrics: allMetrics,
				Error:   err,
			}
		}

	}
	return &RunnerResult{
		Metrics: allMetrics,
		Error:   err,
	}
}
