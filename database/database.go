// Package database provides a database abstraction layer.
package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ORM is the database abstraction layer.
var ORM *gorm.DB

// StartDatabase initializes the database abstraction layer.
func StartDatabase(dbType, dbPath, dbURI string) {
	var err error

	log.Println("Loading database")

	switch dbType {
	case "sqlite":
		ORM, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	case "mysql":
		ORM, err = gorm.Open(mysql.Open(dbURI), &gorm.Config{})
	case "postgresql":
		ORM, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})
	default:
		log.Fatal("Unknown database type")
	}

	if err != nil {
		log.Panic(err)
	}
}
