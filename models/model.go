// Package models contains the database models.
package models

import (
	"github.com/hidracloud/hidra/database"
)

// SetupDB creates the database tables.
func SetupDB(dbType, dbPath, dbURI string) {
	database.StartDatabase(dbType, dbPath, dbURI)

	database.ORM.AutoMigrate(&User{})
	database.ORM.AutoMigrate(&Permission{})
	database.ORM.AutoMigrate(&Agent{})
	database.ORM.AutoMigrate(&AgentTag{})
	database.ORM.AutoMigrate(&Sample{})
	database.ORM.AutoMigrate(&SampleResult{})
	database.ORM.AutoMigrate(&Metric{})
	database.ORM.AutoMigrate(&MetricLabel{})
}
