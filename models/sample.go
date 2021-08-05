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

// Represent a sample, that could be saved on database.
type Sample struct {
	gorm.Model `json:"-"`
	ID         uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name       string    `json:"-"`
	OwnerId    uuid.UUID `json:"-"`
	Owner      User      `gorm:"foreignKey:OwnerId" json:"-"`
	SampleData []byte    `json:"-"`
	Checksum   string
}

// Represent sample metric
type SampleResult struct {
	gorm.Model `json:"-"`
	ID         uuid.UUID `gorm:"primaryKey;type:char(36);"`
	SampleId   uuid.UUID `json:"-"`
	Sample     Sample    `gorm:"foreignKey:SampleId" json:"-"`
	StartDate  time.Time
	EndDate    time.Time
	Error      string
	Agent      Agent     `gorm:"foreignKey:AgentId" json:"-"`
	AgentId    uuid.UUID `json:"-"`
}

// Return a list of samples
func GetSamples() ([]Sample, error) {
	samples := make([]Sample, 0)

	if result := database.ORM.Find(&samples); result.Error != nil {
		return nil, result.Error
	}

	return samples, nil
}

// Get one sample by id
func GetSampleById(id string) (*Sample, error) {
	sample := Sample{}

	if result := database.ORM.First(&sample, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	return &sample, nil
}

// Push metrics to db.
func (s *Sample) PushMetrics(ScenarioResult *ScenarioResult, agentId string) error {

	sample_result := SampleResult{
		ID:        uuid.NewV4(),
		Sample:    *s,
		StartDate: ScenarioResult.StartDate,
		EndDate:   ScenarioResult.EndDate,
		Error:     ScenarioResult.ErrorString,
		AgentId:   uuid.FromStringOrNil(agentId),
	}

	if result := database.ORM.Create(&sample_result); result.Error != nil {
		return result.Error
	}

	common_labels := map[string]string{
		"agent_id":    agentId,
		"sample_id":   s.ID.String(),
		"sample_name": s.Name,
		"checksum":    s.Checksum,
	}

	sample_metric_time := Metric{
		ID:             uuid.NewV4(),
		Name:           "sample_metric_time",
		Value:          float64(sample_result.EndDate.UnixNano() - sample_result.StartDate.UnixNano()),
		LabelsChecksum: CalculateLabelsChecksum(common_labels),
	}

	err := sample_metric_time.PushToDB(common_labels)

	if err != nil {
		return err
	}

	status := 1

	if ScenarioResult.ErrorString != "" {
		status = 0
	}

	sample_metric_status := Metric{
		ID:             uuid.NewV4(),
		Name:           "sample_metric_status",
		Value:          float64(status),
		LabelsChecksum: CalculateLabelsChecksum(common_labels),
	}

	err = sample_metric_status.PushToDB(common_labels)

	if err != nil {
		return err
	}

	for _, step := range ScenarioResult.StepResults {

		common_labels["step"] = step.Step.Type
		common_labels["negate"] = strconv.FormatBool(step.Step.Negate)
		paramsStr, _ := json.Marshal(step.Step.Params)
		common_labels["params"] = string(paramsStr)

		step_sample_metric_time := Metric{
			ID:             uuid.NewV4(),
			Name:           "sample_step_metric_time",
			Value:          float64(step.EndDate.UnixNano() - step.StartDate.UnixNano()),
			LabelsChecksum: CalculateLabelsChecksum(common_labels),
		}

		err = step_sample_metric_time.PushToDB(common_labels)

		if err != nil {
			return err
		}

		for _, metric := range step.Metrics {
			labels := make(map[string]string)
			for k, v := range common_labels {
				labels[k] = v
			}
			for k, v := range metric.Labels {
				labels[k] = v
			}

			metric.ID = uuid.NewV4()
			metric.LabelsChecksum = CalculateLabelsChecksum(labels)
			metric.PushToDB(labels)
		}
	}

	return nil
}

// Register a new sample
func RegisterSample(name string, sampleData []byte, user *User) error {
	checksum := md5.Sum(sampleData)

	newSample := Sample{ID: uuid.NewV4(), Name: name, Owner: *user, SampleData: sampleData, Checksum: hex.EncodeToString(checksum[:])}

	if result := database.ORM.Create(&newSample); result.Error != nil {
		return result.Error
	}

	return nil
}
