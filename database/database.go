// Package database provides a database abstraction layer.
package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB_TYPE string
var DB_PATH string
var DB_URI string

func GetORM(slave bool) (*gorm.DB, error) {
	switch DB_TYPE {
	case "sqlite":
		return gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{})
	case "mysql":
		return gorm.Open(mysql.Open(DB_URI), &gorm.Config{})
	case "postgresql":
		return gorm.Open(postgres.Open(DB_URI), &gorm.Config{})
	}

	return nil, fmt.Errorf("unknow db type %s", DB_TYPE)

}

// StartDatabase initializes the database abstraction layer.
func StartDatabase(dbType, dbPath, dbURI string) {
	DB_TYPE = dbType
	DB_PATH = dbPath
	DB_URI = dbURI
}
