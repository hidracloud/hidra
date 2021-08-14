package models

import (
	"fmt"
	"sort"
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
	Description    string
	Expires        time.Duration
}

// Metric labels
type MetricLabel struct {
	gorm.Model
	Key      string
	Value    string
	Metric   Metric `gorm:"foreignKey:MetricId" json:"-"`
	MetricId string
}

// Delete all expired metrics
func DeleteExpiredMetrics() error {
	if result := database.ORM.Where("expires < ? and expires != 0", time.Now().Unix()).Unscoped().Delete(&Metric{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete old metrics
func DeleteOldMetrics(interval time.Duration) error {
	expiretime := time.Now()
	expiretime.Add(interval)

	if result := database.ORM.Where("updated_at < ?", expiretime.Unix()).Unscoped().Delete(&Metric{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// Get disticnt metric name
func GetDistinctMetricName() ([]string, error) {
	var results []string

	if result := database.ORM.Model(&Metric{}).Distinct().Pluck("name", &results); result.Error != nil {
		return nil, result.Error
	}

	return results, nil
}

// Get distint checksum by name
func GetDistinctChecksumByName(name string) ([]string, error) {
	var results []string
	if result := database.ORM.Model(&Metric{}).Where("name = ?", name).Distinct().Pluck("labels_checksum", &results); result.Error != nil {
		return nil, result.Error
	}
	return results, nil
}

// Get one metric by name
func GetMetricByName(name string) (*Metric, error) {
	var result Metric
	if result := database.ORM.Model(&Metric{}).Where("name = ?", name).Last(&result); result.Error != nil {
		return nil, result.Error
	}

	return &result, nil
}

// Get one metric by checksum
func GetMetricByChecksum(checksum string) (*Metric, error) {
	var result Metric
	if result := database.ORM.Model(&Metric{}).Where("labels_checksum = ?", checksum).Last(&result); result.Error != nil {
		return nil, result.Error
	}
	return &result, nil
}

// Get metric label by metric id
func GetMetricLabelByMetricID(id uuid.UUID) ([]MetricLabel, error) {
	var results []MetricLabel
	if result := database.ORM.Model(&MetricLabel{}).Where("metric_id = ?", id).Find(&results); result.Error != nil {
		return nil, result.Error
	}
	return results, nil
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

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		checksum += k + labels[k]
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
