package models

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var defaultStepTimeout = 15 * time.Second

type stepFn func(map[string]string) ([]Metric, error)

// Step definition
type Step struct {
	Type   string
	Params map[string]string
	Negate bool
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

// Scenarios represent sample scenarios
type Scenarios struct {
	Name        string
	Description string

	Scenario       Scenario
	ScrapeInterval time.Duration `yaml:"scrapeInterval"`
}

// Metric definition
type Metric struct {
	gorm.Model
	ID             uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name           string
	Value          float64
	Labels         map[string]string `gorm:"-"`
	Description    string
	SampleID       string
	Expires        time.Duration
	SampleResultID uuid.UUID     `json:"-"`
	SampleResult   *SampleResult `gorm:"foreignKey:SampleResultID" json:"-"`
}

// MetricLabel definition
type MetricLabel struct {
	gorm.Model
	Key      string
	Value    string
	Metric   Metric `gorm:"foreignKey:MetricID" json:"-"`
	MetricID string
}

// DeleteExpiredMetrics Delete all expired metrics
func DeleteExpiredMetrics() error {
	if result := database.ORM.Where("expires < ? and expires != 0", time.Now().Unix()).Unscoped().Delete(&Metric{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// CleanupMetrics Cleanup metrics
func CleanupMetrics(interval time.Duration) error {
	if err := DeleteOldMetrics(interval); err != nil {
		return err
	}
	if err := DeleteExpiredMetrics(); err != nil {
		return err
	}
	if err := DeleteOldScenarioResults(interval); err != nil {
		return err
	}
	if err := DeleteOldMetricsLabels(interval); err != nil {
		return err
	}
	return nil
}

// DeleteOldMetricsLabels Delete old metrics labels
func DeleteOldMetricsLabels(interval time.Duration) error {
	if result := database.ORM.Where("updated_at < ?", time.Now().Add(-interval).Unix()).Unscoped().Delete(&MetricLabel{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteOldScenarioResults Delete old scenario results
func DeleteOldScenarioResults(interval time.Duration) error {
	if result := database.ORM.Where("updated_at < ?", time.Now().Add(-interval).Unix()).Unscoped().Delete(&ScenarioResult{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteOldMetrics Delete old metrics
func DeleteOldMetrics(interval time.Duration) error {
	expiretime := time.Now()
	expiretime.Add(-interval)

	if result := database.ORM.Where("updated_at < ?", expiretime.Unix()).Unscoped().Delete(&Metric{}); result.Error != nil {
		return result.Error
	}
	return nil
}

// GetDistinctMetricName Get distinct metric name
func GetDistinctMetricName() ([]string, error) {
	var results []string

	if result := database.ORM.Model(&Metric{}).Distinct().Pluck("name", &results); result.Error != nil {
		return nil, result.Error
	}

	return results, nil
}

// GetDistinctChecksumByName Get distinct checksum by name
func GetDistinctChecksumByName(name string) ([]string, error) {
	var results []string
	if result := database.ORM.Model(&Metric{}).Where("name = ?", name).Distinct().Pluck("labels_checksum", &results); result.Error != nil {
		return nil, result.Error
	}
	return results, nil
}

// GetMetricByName Get one metric by name
func GetMetricByName(name string) (*Metric, error) {
	var result Metric
	if result := database.ORM.Model(&Metric{}).Where("name = ?", name).Last(&result); result.Error != nil {
		return nil, result.Error
	}

	return &result, nil
}

// GetMetricByChecksum Get one metric by checksum
func GetMetricByChecksum(checksum, name string) (*Metric, error) {
	var result Metric
	if result := database.ORM.Model(&Metric{}).Where("labels_checksum = ? and name = ?", checksum, name).Last(&result); result.Error != nil {
		return nil, result.Error
	}
	return &result, nil
}

// GetMetricLabelByMetricID Get metric label by metric id
func GetMetricLabelByMetricID(id uuid.UUID) ([]MetricLabel, error) {
	var results []MetricLabel
	if result := database.ORM.Model(&MetricLabel{}).Where("metric_id = ?", id).Find(&results); result.Error != nil {
		return nil, result.Error
	}
	return results, nil
}

// GetDistinctChecksumByNameAndSampleID Get distinct metrics by name and sample id
func GetDistinctChecksumByNameAndSampleID(name, sampleID string) ([]string, error) {
	var results []string
	if result := database.ORM.Model(&Metric{}).Where("name = ? and sample_id = ?", name, sampleID).Distinct().Pluck("labels_checksum", &results); result.Error != nil {
		return nil, result.Error
	}
	return results, nil
}

// GetMetricsByNameAndSampleID Get metrics by name and sample id
func GetMetricsByNameAndSampleID(name, sampleID, checksum string, limit int) ([]Metric, error) {
	var results []Metric
	if result := database.ORM.Model(&Metric{}).Where("name = ? and sample_id = ? and labels_checksum = ?", name, sampleID, checksum).Order("created_at desc").Limit(limit).Find(&results); result.Error != nil {
		return nil, result.Error
	}
	return results, nil
}

// IScenario Define scenario interface
type IScenario interface {
	StartPrimitives()
	Init()
	RunStep(string, map[string]string) ([]Metric, error)
	RegisterStep(string, StepDefinition)
	Description() string
	GetScenarioDefinitions() map[string]StepDefinition
}

// StartPrimitives Initialize primitive variables
func (s *Scenario) StartPrimitives() {
	s.StepsDefinitions = make(map[string]StepDefinition)
}

// PushToDB Push new metric to db
func (m *Metric) PushToDB(labels map[string]string) error {
	if result := database.ORM.Create(m); result.Error != nil {
		return result.Error
	}

	for k, v := range labels {
		label := MetricLabel{
			Key:      k,
			Value:    v,
			MetricID: m.ID.String(),
		}

		if result := database.ORM.Create(&label); result.Error != nil {
			return result.Error
		}
	}
	return nil
}

type runStepGoTemplate struct {
	Now time.Time
}

// RunStep Run an step
func (s *Scenario) RunStep(name string, p map[string]string) ([]Metric, error) {
	if _, ok := s.StepsDefinitions[name]; !ok {
		return nil, fmt.Errorf("sorry but %s not found", name)
	}

	// set runStepGoTemplate
	runStepGoTemplate := &runStepGoTemplate{
		Now: time.Now(),
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

	log.Println("Real parameters: ")
	log.Println(c)

	metricsChain := make(chan []Metric, 1)
	errorChain := make(chan error, 1)

	go func() {
		metrics, err := s.StepsDefinitions[name].Fn(c)
		metricsChain <- metrics
		errorChain <- err
	}()

	select {
	case err := <-errorChain:
		return <-metricsChain, err
	case <-time.After(defaultStepTimeout):
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

// ReadScenariosYAML Read scenarios pointer from yaml
func ReadScenariosYAML(data []byte) (*Scenarios, error) {
	scenarios := Scenarios{}

	err := yaml.Unmarshal([]byte(data), &scenarios)

	if err != nil {
		return nil, err
	}

	return &scenarios, nil
}
