// Package models contains the database models.
package models

import (
	"log"

	"github.com/hidracloud/hidra/database"
)

// SetupDB creates the database tables.
func SetupDB(dbType, dbPath, dbURI string) {
	database.StartDatabase(dbType, dbPath, dbURI)

	orm, err := database.GetORM(false)
	if err != nil {
		log.Fatal(err)
	}

	db, err := orm.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orm.AutoMigrate(&User{})
	orm.AutoMigrate(&Permission{})
	orm.AutoMigrate(&Agent{})
	orm.AutoMigrate(&AgentTag{})
	orm.AutoMigrate(&Sample{})
	orm.AutoMigrate(&SampleResult{})
	orm.AutoMigrate(&Metric{})
	orm.AutoMigrate(&MetricLabel{})
}
