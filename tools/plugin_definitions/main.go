package main

import (
	"encoding/json"

	"github.com/hidracloud/hidra/v3/internal/plugins"
	_ "github.com/hidracloud/hidra/v3/internal/plugins/all"
)

type Plugin2Dump struct {
	Name            string                             `json:"name"`
	Description     string                             `json:"description"`
	StepDefinitions map[string]*plugins.StepDefinition `json:"step_definitions"`
}

func main() {

	allPlugins := plugins.GetPlugins()

	plugins2dump := make([]Plugin2Dump, 0)

	for name := range allPlugins {
		plugin2Dump := Plugin2Dump{
			Name:            name,
			Description:     plugins.GetPluginDescription(name),
			StepDefinitions: plugins.GetPlugin(name).GetSteps(),
		}

		plugins2dump = append(plugins2dump, plugin2Dump)
	}

	// Convert to JSON
	// Print to stdout
	dumpJSON, _ := json.Marshal(plugins2dump)
	println(string(dumpJSON))
}
