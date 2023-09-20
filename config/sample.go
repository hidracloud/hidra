package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// SampleConfig is the sample configuration.
type SampleConfig struct {
	// Name is the sample name.
	Name string `yaml:"name,omitempty"`

	// Path is the sample path.
	Path string `yaml:"path,omitempty"`

	// Description is the description of the sample.
	Description string `yaml:"description,omitempty"`

	// Tags is the tags of the sample.
	Tags map[string]string `yaml:"tags,omitempty"`

	// ScrapeInterval is the interval to scrape the sample.
	Interval time.Duration `yaml:"interval,omitempty"`

	// Timeout is the timeout to scrape the sample.
	Timeout time.Duration `yaml:"timeout,omitempty"`

	// Retry is the retry to scrape the sample.
	Retry int `yaml:"retry,omitempty" default:"0"`

	// Steps is the steps to scrape the sample.
	Steps []StepConfig `yaml:"steps,omitempty"`

	// Variables is the variables to scrape the sample
	Variables []map[string]string `yaml:"variables,omitempty"`
}

// StepConfig is the step configuration.
type StepConfig struct {
	// Plugin is the plugin to scrape the sample. If not value given, the latest used plugin will be used.
	Plugin string `yaml:"plugin,omitempty"`
	// Action is the action to scrape the sample
	Action string `yaml:"action,omitempty"`
	// Parameters is the parameters to scrape the sample
	Parameters map[string]string `yaml:"parameters,omitempty"`
	// Negate is the negate to scrape the sample
	Negate bool `yaml:"negate,omitempty"`
}

// LoadSampleConfig loads from byte array.
func LoadSampleConfig(data []byte) (*SampleConfig, error) {
	var config SampleConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	if config.Variables == nil {
		config.Variables = make([]map[string]string, 0)
	}

	if len(config.Variables) == 0 {
		config.Variables = append(config.Variables, make(map[string]string))
	}

	if config.Interval == 0 {
		config.Interval = 60 * time.Second
	}

	if config.Retry < 0 {
		config.Retry = 0
	}

	return &config, nil
}

// LoadSampleConfigFromFile loads from file.
func LoadSampleConfigFromFile(path string) (*SampleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cnf, err := LoadSampleConfig(data)

	if err != nil {
		return nil, err
	}

	cnf.Path = path

	return cnf, nil
}

// Verify verifies the sample configuration.
func (c *SampleConfig) Verify() error {
	if c.Description == "" {
		return fmt.Errorf("description is required")
	}

	for _, step := range c.Steps {
		if step.Action == "" {
			return fmt.Errorf("action is required")
		}

		if step.Parameters == nil {
			return fmt.Errorf("parameters is required")
		}
	}

	return nil
}
