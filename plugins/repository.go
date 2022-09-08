package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/hidracloud/hidra/v3/internal/config"
	"github.com/hidracloud/hidra/v3/internal/metrics"
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

// RunSample runs a sample.
func RunSample(ctx context.Context, sample *config.SampleConfig) (context.Context, []*metrics.Metric, error) {
	var newMetrics []*metrics.Metric

	allMetrics := []*metrics.Metric{}

	depthSize := 1

	lastPlugin := ""

	for _, step := range sample.Steps {
		log.Debugf("|%s Running plugin %s", strings.Repeat("_", depthSize), step.Plugin)
		log.Debugf("|_%s Action: %v", strings.Repeat("_", depthSize), step.Action)
		log.Debugf("|_%s Parameters: ", strings.Repeat("_", depthSize))

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
		var err error
		ctx, newMetrics, err = plugin.RunStep(ctx, &Step{
			Name: step.Action,
			Args: step.Parameters,
		})

		allMetrics = append(allMetrics, newMetrics...)

		if err != nil {
			return ctx, allMetrics, err
		}
	}
	return ctx, allMetrics, nil
}
