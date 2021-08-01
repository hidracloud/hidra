// Represent a data model
package models

import (
	"log"

	"github.com/JoseCarlosGarcia95/hidra/database"
	"github.com/JoseCarlosGarcia95/hidra/utils"
	uuid "github.com/satori/go.uuid"
)

func SetupDB() {
	database.StartDatabase()

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
