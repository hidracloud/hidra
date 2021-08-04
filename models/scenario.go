package models

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

type stepFn func(map[string]string) ([]CustomMetric, error)

// Define one step
type Step struct {
	Type   string
	Params map[string]string
	Negate bool
}

// Define one scenario
type Scenario struct {
	Kind    string
	Steps   []Step
	StepsFn map[string]stepFn
}

// Define step metrics
type StepMetric struct {
	Step      Step
	StartDate time.Time
	EndDate   time.Time
}

// Define scenario metrics
type ScenarioMetric struct {
	Scenario    Scenario
	StartDate   time.Time
	EndDate     time.Time
	StepMetrics []*StepMetric
	Error       error
	ErrorString string
}

// Define a set of scenarios
type Scenarios struct {
	Name        string
	Description string

	Scenario       Scenario
	ScrapeInterval time.Duration `yaml:"scrapeInterval"`
}

// Custom metric for scenarios
type CustomMetric struct {
	Name   string
	Value  float64
	Labels map[string]string
}

// Define scenario interface
type IScenario interface {
	StartPrimitives()
	Init()
	RunStep(string, map[string]string) ([]CustomMetric, error)
	RegisterStep(string, stepFn)
}

// Initialize primitive variables
func (s *Scenario) StartPrimitives() {
	s.StepsFn = make(map[string]stepFn)
}

// Run an step
func (s *Scenario) RunStep(name string, c map[string]string) ([]CustomMetric, error) {
	if _, ok := s.StepsFn[name]; !ok {
		return nil, fmt.Errorf("sorry but %s not found", name)
	}
	return s.StepsFn[name](c)
}

// Register step default method
func (s *Scenario) RegisterStep(name string, step stepFn) {
	s.StepsFn[name] = step
}

// Read scenarios pointer from yaml
func ReadScenariosYAML(data []byte) (*Scenarios, error) {
	scenarios := Scenarios{}

	err := yaml.Unmarshal([]byte(data), &scenarios)

	if err != nil {
		return nil, err
	}

	return &scenarios, nil
}
