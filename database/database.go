// Represent a database connection
package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ORM *gorm.DB

func StartDatabase() {
	var err error

	log.Println("Loading database")
	ORM, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	if err != nil {
		log.Panic(err)
	}

	log.Println("Loading database models")
}
