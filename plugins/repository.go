package plugins

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/utils"
	log "github.com/sirupsen/logrus"
)

var (
	plugins = make(map[string]PluginInterface)
)

// AddPlugin adds a plugin.
func AddPlugin(name string, plugin PluginInterface) {
	plugins[name] = plugin
}

// GetPlugin returns a plugin.
func GetPlugin(name string) PluginInterface {
	return plugins[name]
}

// StepParamTemplate represents a step parameter template.
type StepParamTemplate struct {
	Env       map[string]string
	Date      time.Time
	Context   context.Context
	Variables map[string]string
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
func RunWithVariables(ctx context.Context, variables map[string]string, sample *config.SampleConfig) (context.Context, []*metrics.Metric, error) {
	var allMetrics, newMetrics []*metrics.Metric

	depthSize := 1
	lastPlugin := ""
	pluginsByNames := make(map[string]PluginInterface)

	stepParamTemplate := StepParamTemplate{
		Env:     utils.EnvToMap(),
		Date:    time.Now(),
		Context: ctx,
	}

	variables, err := stepParamTemplate.Replace(variables)

	if err != nil {
		return ctx, nil, err
	}

	stepParamTemplate.Variables = variables

	for _, step := range sample.Steps {
		log.Debugf("|%s Running plugin %s", strings.Repeat("_", depthSize), step.Plugin)
		log.Debugf("|_%s Action: %v", strings.Repeat("_", depthSize), step.Action)
		log.Debugf("|_%s Parameters: ", strings.Repeat("_", depthSize))

		stepParamTemplate.Context = ctx

		step.Parameters, err = stepParamTemplate.Replace(step.Parameters)

		if err != nil {
			return ctx, nil, err
		}

		for k, v := range step.Parameters {
			log.Debugf("|__%s %s: %v", strings.Repeat("_", depthSize), k, v)
		}

		depthSize++
		if step.Plugin == "" {
			step.Plugin = lastPlugin
		}
		plugin := GetPlugin(step.Plugin)

		if plugin == nil {
			return ctx, allMetrics, fmt.Errorf("plugin %s not found", step.Plugin)
		}

		lastPlugin = step.Plugin
		pluginsByNames[step.Plugin] = plugin

		ctx, newMetrics, err = plugin.RunStep(ctx, &Step{
			Name:   step.Action,
			Args:   step.Parameters,
			Negate: step.Negate,
		})

		allMetrics = append(allMetrics, newMetrics...)

		if err != nil {
			return ctx, allMetrics, err
		}
	}

	// Clean up plugins
	for _, plugin := range pluginsByNames {
		if plugin.StepExists("onClose") {
			_, _, err = plugin.RunStep(ctx, &Step{
				Name: "onClose",
				Args: map[string]string{},
			})

			if err != nil {
				log.Warnf("Error closing plugin: %v", err)
				err = nil
			}
		}
	}

	return ctx, allMetrics, err
}

// RunSample runs a sample.
func RunSample(ctx context.Context, sample *config.SampleConfig) (context.Context, []*metrics.Metric, error) {
	var err error

	allMetrics := []*metrics.Metric{}

	ctx = context.WithValue(ctx, ContextTimeout, sample.Timeout)

	for _, variables := range sample.Variables {
		ctx, newMetrics, err := RunWithVariables(ctx, variables, sample)

		allMetrics = append(allMetrics, newMetrics...)

		if err != nil {
			return ctx, allMetrics, err
		}

	}
	return ctx, allMetrics, err
}
