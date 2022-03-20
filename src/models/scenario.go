package models

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/hidracloud/hidra/src/utils"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
)

var envMap map[string]string

var minStepTimeout = 10 * time.Second

type stepFn func(map[string]string) ([]Metric, error)

// Step definition
type Step struct {
	Type    string
	Params  map[string]string
	Negate  bool
	Timeout time.Duration
}

// StepParam returns the value of a step parameter
type StepParam struct {
	Name        string
	Description string
	Optional    bool
}

// StepDefinition definition
type StepDefinition struct {
	Name        string
	Description string
	Params      []StepParam
	Fn          stepFn `json:"-"`
}

// StepResult is the result of a step
type StepResult struct {
	Step      Step
	StartDate time.Time
	EndDate   time.Time
	Metrics   []Metric
}

// Scenario definition
type Scenario struct {
	Kind             string
	Steps            []Step
	StepsDefinitions map[string]StepDefinition
	Context          context.Context
}

// ScenarioResult is the result of a scenario
type ScenarioResult struct {
	Scenario    Scenario
	StartDate   time.Time
	EndDate     time.Time
	StepResults []*StepResult
	Error       error `json:"-"`
	ErrorString string
}

// Sample represent sample scenarios
type Sample struct {
	Name           string
	Description    string
	Tags           map[string]string
	Scenario       Scenario
	ScrapeInterval time.Duration `yaml:"scrapeInterval"`
}

// Metric definition
type Metric struct {
	ID             uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name           string
	Value          float64
	Labels         map[string]string `gorm:"-"`
	Description    string
	SampleID       string
	Expires        time.Duration
	SampleResultID uuid.UUID `json:"-"`
}

// MetricLabel definition
type MetricLabel struct {
	Key      string
	Value    string
	Metric   Metric `gorm:"foreignKey:MetricID" json:"-"`
	MetricID string
}

// IScenario Define scenario interface
type IScenario interface {
	StartPrimitives()
	Init()
	Close()
	RunStep(string, map[string]string, time.Duration) ([]Metric, error)
	RegisterStep(string, StepDefinition)
	Description() string
	GetScenarioDefinitions() map[string]StepDefinition
}

// StartPrimitives Initialize primitive variables
func (s *Scenario) StartPrimitives() {
	s.StepsDefinitions = make(map[string]StepDefinition)
}

type runStepGoTemplate struct {
	Now time.Time
	Env map[string]string
}

// RunStep Run an step
func (s *Scenario) RunStep(name string, p map[string]string, timeout time.Duration) ([]Metric, error) {
	if _, ok := s.StepsDefinitions[name]; !ok {
		return nil, fmt.Errorf("sorry but %s not found", name)
	}

	// set runStepGoTemplate
	runStepGoTemplate := &runStepGoTemplate{
		Now: time.Now(),
		Env: envMap,
	}

	// Make a copy of p into c
	c := make(map[string]string)
	for k, v := range p {
		c[k] = v
	}

	params := s.StepsDefinitions[name].Params

	for _, param := range params {
		if _, ok := c[param.Name]; !ok && !param.Optional {
			return nil, fmt.Errorf("missing parameter %s but expected", param.Name)
		}

		// Parse c[param.Name] as golang template
		t, err := template.New("").Parse(c[param.Name])
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, runStepGoTemplate)
		if err != nil {
			return nil, err
		}

		c[param.Name] = buf.String()
	}

	utils.LogDebug("Running step %s with params %v\n", name, c)

	metricsChain := make(chan []Metric, 1)
	errorChain := make(chan error, 1)

	go func() {
		metrics, err := s.StepsDefinitions[name].Fn(c)
		metricsChain <- metrics
		errorChain <- err
	}()

	if timeout < minStepTimeout {
		timeout = minStepTimeout
	}

	select {
	case err := <-errorChain:
		metrics := <-metricsChain
		close(metricsChain)
		close(errorChain)
		return metrics, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("your step generated a timeout")
	}
}

// RegisterStep Register step default method
func (s *Scenario) RegisterStep(name string, step StepDefinition) {
	s.StepsDefinitions[name] = step
}

// GetScenarioDefinitions Get scenario definitions
func (s *Scenario) GetScenarioDefinitions() map[string]StepDefinition {
	return s.StepsDefinitions
}

// Description Get scenario description
func (s *Scenario) Description() string {
	return ""
}

// ReadSampleYAML Read scenarios pointer from yaml
func ReadSampleYAML(data []byte) (*Sample, error) {
	scenarios := Sample{}

	err := yaml.Unmarshal([]byte(data), &scenarios)

	if err != nil {
		return nil, err
	}

	return &scenarios, nil
}

func init() {
	var err error

	envMap, err = utils.EnvToMap()

	if err != nil {
		panic(err)
	}
}
