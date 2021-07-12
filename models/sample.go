package models

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/JoseCarlosGarcia95/hidra/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Sample struct {
	gorm.Model `json:"-"`
	ID         uuid.UUID `gorm:"primaryKey;type:char(36);"`
	Name       string    `json:"-"`
	OwnerId    uuid.UUID `json:"-"`
	Owner      User      `gorm:"foreignKey:OwnerId" json:"-"`
	SampleData []byte    `json:"-"`
	Checksum   string
}

func GetSamples() ([]Sample, error) {
	samples := make([]Sample, 0)

	if result := database.ORM.Find(&samples); result.Error != nil {
		return nil, result.Error
	}

	return samples, nil
}

func GetSampleById(id string) (*Sample, error) {
	sample := Sample{}

	if result := database.ORM.First(&sample, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	return &sample, nil
}

func RegisterSample(name string, sampleData []byte, user *User) error {
	checksum := md5.Sum(sampleData)

	newSample := Sample{ID: uuid.NewV4(), Name: name, Owner: *user, SampleData: sampleData, Checksum: hex.EncodeToString(checksum[:])}

	if result := database.ORM.Create(&newSample); result.Error != nil {
		return result.Error
	}

	return nil
}
