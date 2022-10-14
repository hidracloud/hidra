package main

import (
	"encoding/json"
	"os"

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

	for _, plugin := range plugins2dump {
		readmeTxt := "# " + plugin.Name + "\n" + plugin.Description + "\n## Available actions\n"
		for _, step := range plugin.StepDefinitions {
			readmeTxt += "### " + step.Name + "\n" + step.Description + "\n"
			readmeTxt += "#### Parameters\n"
			for _, param := range step.Params {
				optional := ""

				if param.Optional {
					optional = " (optional) "
				}
				readmeTxt += "- " + optional + param.Name + ": " + param.Description + "\n"
			}
		}

		// create directory if not exists
		err := os.MkdirAll("docs/plugins/"+plugin.Name, os.ModePerm)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("docs/plugins/"+plugin.Name+"/README.md", []byte(readmeTxt), 0644)
		if err != nil {
			panic(err)
		}
	}
	// Convert to JSON
	// Print to stdout
	dumpJSON, _ := json.Marshal(plugins2dump)
	println(string(dumpJSON))
}
