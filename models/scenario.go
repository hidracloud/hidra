package models

import (
	"fmt"
	"time"

	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

type stepFn func(map[string]string) ([]Metric, error)

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
type StepResult struct {
	Step      Step
	StartDate time.Time
	EndDate   time.Time
	Metrics   []Metric
}

// Define scenario metrics
type ScenarioResult struct {
	Scenario    Scenario
	StartDate   time.Time
	EndDate     time.Time
	StepResults []*StepResult
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

// metric for scenarios
type Metric struct {
	gorm.Model
	ID             uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name           string
	Value          float64
	Step           string
	Scenario       string
	Labels         map[string]string `gorm:"-"`
	LabelsChecksum string
}

// Metric labels
type MetricLabel struct {
	gorm.Model
	Key      string
	Value    string
	Metric   Metric `gorm:"foreignKey:MetricId" json:"-"`
	MetricId string
}

// Define scenario interface
type IScenario interface {
	StartPrimitives()
	Init()
	RunStep(string, map[string]string) ([]Metric, error)
	RegisterStep(string, stepFn)
}

// Initialize primitive variables
func (s *Scenario) StartPrimitives() {
	s.StepsFn = make(map[string]stepFn)
}

// Push new metric to db
func (m *Metric) PushToDB(labels map[string]string) error {
	if result := database.ORM.Create(m); result.Error != nil {
		return result.Error
	}

	for k, v := range labels {
		label := MetricLabel{
			Key:      k,
			Value:    v,
			MetricId: m.ID.String(),
		}

		if result := database.ORM.Create(&label); result.Error != nil {
			return result.Error
		}
	}
	return nil
}

// Run an step
func (s *Scenario) RunStep(name string, c map[string]string) ([]Metric, error) {
	if _, ok := s.StepsFn[name]; !ok {
		return nil, fmt.Errorf("sorry but %s not found", name)
	}
	return s.StepsFn[name](c)
}

// Register step default method
func (s *Scenario) RegisterStep(name string, step stepFn) {
	s.StepsFn[name] = step
}

// Calculate labels checksum
func CalculateLabelsChecksum(labels map[string]string) string {
	var checksum string
	for k, v := range labels {
		checksum += k + v
	}
	return checksum
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
