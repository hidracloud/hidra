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
	Env     map[string]string
	Date    time.Time
	Context context.Context
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

// RunSample runs a sample.
func RunSample(ctx context.Context, sample *config.SampleConfig) (context.Context, []*metrics.Metric, error) {
	var newMetrics []*metrics.Metric
	var err error

	allMetrics := []*metrics.Metric{}
	pluginsByNames := make(map[string]PluginInterface)

	depthSize := 1

	lastPlugin := ""

	ctx = context.WithValue(ctx, ContextTimeout, sample.Timeout)

	stepParamTemplate := StepParamTemplate{
		Env:     utils.EnvToMap(),
		Date:    time.Now(),
		Context: ctx,
	}

	for _, step := range sample.Steps {
		log.Debugf("|%s Running plugin %s", strings.Repeat("_", depthSize), step.Plugin)
		log.Debugf("|_%s Action: %v", strings.Repeat("_", depthSize), step.Action)
		log.Debugf("|_%s Parameters: ", strings.Repeat("_", depthSize))

		stepParamTemplate.Context = ctx

		for k, v := range step.Parameters {
			t, err := template.New("").Parse(v)
			if err != nil {
				return ctx, nil, err
			}

			var buf bytes.Buffer
			err = t.Execute(&buf, stepParamTemplate)
			if err != nil {
				return ctx, nil, err
			}

			v = buf.String()
			step.Parameters[k] = buf.String()
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
			Name: step.Action,
			Args: step.Parameters,
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
			}
		}
	}
	return ctx, allMetrics, err
}
