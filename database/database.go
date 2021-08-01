// Represent a database connection
package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var ORM *gorm.DB

func StartDatabase(db_type, db_path, db_uri string) {
	var err error

	log.Println("Loading database")

	switch db_type {
	case "sqlite":
		ORM, err = gorm.Open(sqlite.Open(db_path), &gorm.Config{})
	case "mysql":
		ORM, err = gorm.Open(mysql.Open(db_uri), &gorm.Config{})
	case "postgresql":
		ORM, err = gorm.Open(postgres.Open(db_uri), &gorm.Config{})
	default:
		log.Fatal("Unknown database type")
	}

	if err != nil {
		log.Panic(err)
	}

	log.Println("Loading database models")
}
