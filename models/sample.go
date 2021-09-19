package models

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hidracloud/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// MetricQueue represent a metric queue
type MetricQueue struct {
	AgentID        string
	Sample         Sample
	ScenarioResult ScenarioResult
}

var metricsQueue []MetricQueue

// Sample represent a sample
type Sample struct {
	gorm.Model  `json:"-"`
	ID          uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name        string    `json:"-"`
	OwnerID     uuid.UUID `json:"-"`
	Owner       User      `gorm:"foreignKey:OwnerID" json:"-"`
	SampleData  []byte    `json:"-"`
	Checksum    string
	Description string `json:"-"`
	Kind        string `json:"-"`
}

// SampleResult represent a sample result
type SampleResult struct {
	gorm.Model `json:"-"`
	ID         uuid.UUID `gorm:"primaryKey;type:char(36);"`
	SampleID   uuid.UUID `json:"-"`
	Sample     Sample    `gorm:"foreignKey:SampleID" json:"-"`
	StartDate  time.Time
	EndDate    time.Time
	Error      string
	Agent      Agent `gorm:"foreignKey:AgentID"`
	AgentID    uuid.UUID
}

// SearchSamplesWithPagination search samples with pagination
func SearchSamplesWithPagination(name string, page, limit int) ([]Sample, error) {
	samples := []Sample{}
	if result := database.ORM.Where("name LIKE ?", "%"+name+"%").Order("updated_at desc").Offset(page * limit).Limit(limit).Find(&samples); result.Error != nil {
		return samples, result.Error
	}
	return samples, nil
}

// GetTotalSamples return total samples
func GetTotalSamples() int64 {
	var count int64
	database.ORM.Model(&Sample{}).Count(&count)
	return count
}

// GetLastSampleResultBySampleID return last sample result by sample id
func GetLastSampleResultBySampleID(sampleID string) (*SampleResult, error) {
	sampleResult := SampleResult{}
	if result := database.ORM.Where("sample_id = ?", sampleID).Order("created_at desc").First(&sampleResult); result.Error != nil {
		return nil, result.Error
	}
	return &sampleResult, nil
}

// SearchSamples search samples
func SearchSamples(name string) ([]Sample, error) {
	samples := []Sample{}
	if result := database.ORM.Order("updated_at desc").Where("name LIKE ?", "%"+name+"%").Find(&samples); result.Error != nil {
		return nil, result.Error
	}
	return samples, nil
}

// GetSamplesWithPagination return samples with pagination
func GetSamplesWithPagination(page, limit int) ([]Sample, error) {
	samples := []Sample{}
	if result := database.ORM.Order("updated_at desc").Offset(page * limit).Limit(limit).Find(&samples); result.Error != nil {
		return nil, result.Error
	}
	return samples, nil
}

// GetSamples return samples
func GetSamples() ([]Sample, error) {
	samples := make([]Sample, 0)

	if result := database.ORM.Order("updated_at desc").Find(&samples); result.Error != nil {
		return nil, result.Error
	}

	return samples, nil
}

// GetSampleResults return sample results by sample id
func GetSampleResults(sampleID string, limit int) ([]SampleResult, error) {
	sampleResults := []SampleResult{}
	if result := database.ORM.Where("sample_id = ?", sampleID).Order("created_at desc").Limit(limit).Preload("Agent").Find(&sampleResults); result.Error != nil {
		return nil, result.Error
	}
	return sampleResults, nil
}

// GetSampleQuery return sample query
func GetSampleQuery() *gorm.DB {
	return database.ORM.Order("updated_at desc")
}

// GetSampleByID return sample by id
func GetSampleByID(id string) (*Sample, error) {
	sample := Sample{}

	if result := database.ORM.First(&sample, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	return &sample, nil
}

// GetMetricsBySampleID return metrics by sample id
func GetMetricsBySampleID(sampleID string) ([]Metric, error) {
	metricResult := []Metric{}
	if result := database.ORM.Where("sample_id = ?", sampleID).Find(&metricResult); result.Error != nil {
		return nil, result.Error
	}
	return metricResult, nil
}

// GetDistinctMetricNameBySampleID return distinct metric name by sample id
func GetDistinctMetricNameBySampleID(sampleID string) ([]string, error) {
	var results []string

	if result := database.ORM.Model(&Metric{}).Distinct().Where("sample_id = ?", sampleID).Pluck("name", &results); result.Error != nil {
		return nil, result.Error
	}

	return results, nil
}

// ProcessMetricsQueue process metrics queue
func ProcessMetricsQueue() {
	for len(metricsQueue) > 0 {
		metricsQueue[0].Sample.PushMetrics(&metricsQueue[0].ScenarioResult, metricsQueue[0].AgentID)
		metricsQueue = metricsQueue[1:]
	}
}

// PushMetricsToQueue push metrics to queue
func (s *Sample) PushMetricsToQueue(ScenarioResult *ScenarioResult, agentID string) {
	metricsQueue = append(metricsQueue, MetricQueue{AgentID: agentID, Sample: *s, ScenarioResult: *ScenarioResult})
}

// GetSampleResultBySampleIDWithLimit return latest sample result by sample id
func GetSampleResultBySampleIDWithLimit(sampleID string, limit int) ([]*SampleResult, error) {
	sampleResult := []*SampleResult{}
	if result := database.ORM.Where("sample_id = ?", sampleID).Order("created_at desc").Limit(limit).Find(&sampleResult); result.Error != nil {
		return nil, result.Error
	}
	return sampleResult, nil
}

// GetMetricsBySampleResultID return metrics by sample result id
func GetMetricsBySampleResultID(sampleResultID string) ([]Metric, error) {
	metricResult := []Metric{}
	if result := database.ORM.Where("sample_result_id = ?", sampleResultID).Find(&metricResult); result.Error != nil {
		return nil, result.Error
	}
	return metricResult, nil
}

// PushMetrics push metrics to database
func (s *Sample) PushMetrics(ScenarioResult *ScenarioResult, agentID string) error {

	sampleResult := SampleResult{
		ID:        uuid.NewV4(),
		Sample:    *s,
		StartDate: ScenarioResult.StartDate,
		EndDate:   ScenarioResult.EndDate,
		Error:     ScenarioResult.ErrorString,
		AgentID:   uuid.FromStringOrNil(agentID),
	}

	if result := database.ORM.Create(&sampleResult); result.Error != nil {
		return result.Error
	}

	agent, err := GetAgent(uuid.FromStringOrNil(agentID))
	if err != nil {
		return err
	}

	commonLabels := map[string]string{
		"agent_id":    agentID,
		"agent_name":  agent.Name,
		"sample_id":   s.ID.String(),
		"sample_name": s.Name,
		"checksum":    s.Checksum,
	}

	sampleMetricTime := Metric{
		ID:           uuid.NewV4(),
		SampleID:     s.ID.String(),
		Name:         "sample_metric_time",
		SampleResult: &sampleResult,
		Value:        float64(sampleResult.EndDate.UnixNano() - sampleResult.StartDate.UnixNano()),
	}

	err = sampleMetricTime.PushToDB(commonLabels)

	if err != nil {
		return err
	}

	status := 1

	if ScenarioResult.ErrorString != "" {
		status = 0
	}

	sampleMetricStatus := Metric{
		ID:           uuid.NewV4(),
		Name:         "sample_metric_status",
		SampleID:     s.ID.String(),
		Value:        float64(status),
		SampleResult: &sampleResult,
	}

	err = sampleMetricStatus.PushToDB(commonLabels)

	if err != nil {
		return err
	}

	for _, step := range ScenarioResult.StepResults {

		commonLabels["step"] = step.Step.Type
		commonLabels["negate"] = strconv.FormatBool(step.Step.Negate)
		paramsStr, _ := json.Marshal(step.Step.Params)
		commonLabels["params"] = string(paramsStr)

		stepSampleMetricTime := Metric{
			ID:           uuid.NewV4(),
			Name:         "sample_step_metric_time",
			SampleID:     s.ID.String(),
			Value:        float64(step.EndDate.UnixNano() - step.StartDate.UnixNano()),
			SampleResult: &sampleResult,
		}

		err = stepSampleMetricTime.PushToDB(commonLabels)

		if err != nil {
			return err
		}

		for _, metric := range step.Metrics {
			labels := make(map[string]string)
			for k, v := range commonLabels {
				labels[k] = v
			}
			for k, v := range metric.Labels {
				labels[k] = v
			}

			metric.ID = uuid.NewV4()
			metric.SampleID = s.ID.String()
			metric.SampleResult = &sampleResult
			metric.PushToDB(labels)
		}
	}

	return nil
}

// UpdateSample update sample
func UpdateSample(id, name, descrption, kind string, sampleData []byte, user *User) (*Sample, error) {
	suuid, err := uuid.FromString(id)

	if err != nil {
		return nil, err
	}

	checksum := md5.Sum(sampleData)
	updateSample := Sample{ID: suuid, Name: name, Kind: kind, Description: descrption, Owner: *user, SampleData: sampleData, Checksum: hex.EncodeToString(checksum[:])}

	// Save the sample
	if result := database.ORM.Save(&updateSample); result.Error != nil {
		return nil, result.Error
	}
	return &updateSample, nil
}

// RegisterSample 	register sample
func RegisterSample(name, descrption, kind string, sampleData []byte, user *User) (*Sample, error) {
	checksum := md5.Sum(sampleData)

	newSample := Sample{ID: uuid.NewV4(), Name: name, Kind: kind, Description: descrption, Owner: *user, SampleData: sampleData, Checksum: hex.EncodeToString(checksum[:])}

	if result := database.ORM.Create(&newSample); result.Error != nil {
		return nil, result.Error
	}

	return &newSample, nil
}

func init() {
	metricsQueue = make([]MetricQueue, 0)
}
