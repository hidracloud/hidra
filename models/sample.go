package models

import (
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/JoseCarlosGarcia95/hidra/database"
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

// Represent a sample step metric
type SampleStepMetric struct {
	gorm.Model     `json:"-"`
	ID             uuid.UUID    `gorm:"primaryKey;type:char(36);"`
	SampleMetricId uuid.UUID    `json:"-"`
	SampleMetric   SampleMetric `gorm:"foreignKey:SampleMetricId" json:"-"`
	StartDate      time.Time
	EndDate        time.Time
}

// Represent sample metric
type SampleMetric struct {
	gorm.Model `json:"-"`
	ID         uuid.UUID `gorm:"primaryKey;type:char(36);"`
	SampleId   uuid.UUID `json:"-"`
	Sample     Sample    `gorm:"foreignKey:SampleId" json:"-"`
	StartDate  time.Time
	EndDate    time.Time
	Error      string
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
func (s *Sample) PushMetrics(scenarioMetric *ScenarioMetric) error {
	sampleMetric := SampleMetric{
		ID:        uuid.NewV4(),
		Sample:    *s,
		StartDate: scenarioMetric.StartDate,
		EndDate:   scenarioMetric.EndDate,
		Error:     scenarioMetric.ErrorString,
	}

	if result := database.ORM.Create(&sampleMetric); result.Error != nil {
		return result.Error
	}

	for _, step := range scenarioMetric.StepMetrics {
		sampleStepMetric := SampleStepMetric{
			ID:           uuid.NewV4(),
			StartDate:    step.StartDate,
			EndDate:      step.EndDate,
			SampleMetric: sampleMetric,
		}

		if result := database.ORM.Create(&sampleStepMetric); result.Error != nil {
			return result.Error
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
