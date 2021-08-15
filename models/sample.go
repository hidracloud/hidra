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

// Represent push metric queue
type MetricQueue struct {
	AgentId        string
	Sample         Sample
	ScenarioResult ScenarioResult
}

var metricsQueue []MetricQueue

// Represent a sample, that could be saved on database.
type Sample struct {
	gorm.Model  `json:"-"`
	ID          uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name        string    `json:"-"`
	OwnerId     uuid.UUID `json:"-"`
	Owner       User      `gorm:"foreignKey:OwnerId" json:"-"`
	SampleData  []byte    `json:"-"`
	Checksum    string
	Description string `json:"-"`
	Kind        string `json:"-"`
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

// Search samples by name
func SearchSamples(name string) ([]Sample, error) {
	samples := []Sample{}
	if result := database.ORM.Order("updated_at desc").Where("name LIKE ?", "%"+name+"%").Find(&samples); result.Error != nil {
		return nil, result.Error
	}
	return samples, nil
}

// Return a list of samples
func GetSamples() ([]Sample, error) {
	samples := make([]Sample, 0)

	if result := database.ORM.Order("updated_at desc").Find(&samples); result.Error != nil {
		return nil, result.Error
	}

	return samples, nil
}

// Get common sample query
func GetSampleQuery() *gorm.DB {
	return database.ORM.Order("updated_at desc")
}

// Get one sample by id
func GetSampleById(id string) (*Sample, error) {
	sample := Sample{}

	if result := database.ORM.First(&sample, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	return &sample, nil
}

// Process metrics queue
func ProcessMetricsQueue() {
	for len(metricsQueue) > 0 {
		metricsQueue[0].Sample.PushMetrics(&metricsQueue[0].ScenarioResult, metricsQueue[0].AgentId)
		metricsQueue = metricsQueue[1:]
	}
}

// Push metrics to queue
func (s *Sample) PushMetricsToQueue(ScenarioResult *ScenarioResult, agentId string) {
	metricsQueue = append(metricsQueue, MetricQueue{AgentId: agentId, Sample: *s, ScenarioResult: *ScenarioResult})
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

	agent, err := GetAgent(uuid.FromStringOrNil(agentId))
	if err != nil {
		return err
	}

	common_labels := map[string]string{
		"agent_id":    agentId,
		"agent_name":  agent.Name,
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

	err = sample_metric_time.PushToDB(common_labels)

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

// Update a sample
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

// Register a new sample
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
