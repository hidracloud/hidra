// Represent a data model
package models

import (
	"log"

	"github.com/hidracloud/hidra/database"
	"github.com/hidracloud/hidra/utils"
	uuid "github.com/satori/go.uuid"
)

func SetupDB(db_type, db_path, db_uri string) {
	database.StartDatabase(db_type, db_path, db_uri)

	database.ORM.AutoMigrate(&User{})
	database.ORM.AutoMigrate(&Permission{})
	database.ORM.AutoMigrate(&Agent{})
	database.ORM.AutoMigrate(&AgentTag{})
	database.ORM.AutoMigrate(&Sample{})
	database.ORM.AutoMigrate(&SampleMetric{})
	database.ORM.AutoMigrate(&SampleStepMetric{})

	admin := GetUserByEmail("root")
	if admin.ID == uuid.Nil {
		randomPass := utils.RandString(32)
		log.Println("Creating admin Account for first setup with pass:", randomPass)
		user, err := CreateUser("root", randomPass)

		if err != nil {
			log.Fatal(err)
		}

		AddPermission2User(user, "superadmin")
	}
}
