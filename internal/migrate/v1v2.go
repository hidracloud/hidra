package migrate

import (
	"os"
	"time"

	"github.com/hidracloud/hidra/v3/internal/config"
	"gopkg.in/yaml.v3"
)

// SampleV1V2 represent sample scenarios
type SampleV1V2 struct {
	Name           string
	Description    string
	Tags           map[string]string
	Scenario       ScenarioV1V2
	ScrapeInterval time.Duration `yaml:"scrapeInterval"`
}

// ScenarioV1V2 definition
type ScenarioV1V2 struct {
	Kind             string
	Steps            []StepV1V2
	StepsDefinitions map[string]StepDefinitionV1V2
}

// StepV1V2 definition
type StepV1V2 struct {
	Type    string
	Params  map[string]string
	Negate  bool
	Timeout time.Duration
}

// StepParamV1V2 returns the value of a step parameter
type StepParamV1V2 struct {
	Name        string
	Description string
	Optional    bool
}

// StepDefinitionV1V2 definition
type StepDefinitionV1V2 struct {
	Name        string
	Description string
	Params      []StepParamV1V2
}

// LoadSampleV1V2Config loads from byte array.
func LoadSampleV1V2Config(data []byte) (*SampleV1V2, error) {
	var config SampleV1V2
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadSampleV1V2ConfigFromFile loads from file.
func LoadSampleV1V2ConfigFromFile(path string) (*SampleV1V2, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadSampleV1V2Config(data)
}

// Migrate converts a v1-v2 sample to a v3 sample
func (s *SampleV1V2) Migrate() *config.SampleConfig {
	sample := &config.SampleConfig{
		Name:        s.Name,
		Description: s.Description,
		Tags:        s.Tags,
		Interval:    s.ScrapeInterval,
		Timeout:     60 * time.Second,
		Variables:   make([]map[string]string, 0),
	}

	for _, step := range s.Scenario.Steps {
		if step.Type == "dumpMetrics" {
			continue
		}

		params := make(map[string]string)

		for k, v := range step.Params {
			if k == "test-file" {
				k = "file"
			}
			params[k] = v
		}

		sample.Steps = append(sample.Steps, config.StepConfig{
			Plugin:     s.Scenario.Kind,
			Action:     step.Type,
			Parameters: params,
			Negate:     step.Negate,
		})
	}

	return sample
}
