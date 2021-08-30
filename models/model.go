// Represent a data model
package models

import (
	"github.com/hidracloud/hidra/database"
)

func SetupDB(db_type, db_path, db_uri string) {
	database.StartDatabase(db_type, db_path, db_uri)

	database.ORM.AutoMigrate(&User{})
	database.ORM.AutoMigrate(&Permission{})
	database.ORM.AutoMigrate(&Agent{})
	database.ORM.AutoMigrate(&AgentTag{})
	database.ORM.AutoMigrate(&Sample{})
	database.ORM.AutoMigrate(&SampleResult{})
	database.ORM.AutoMigrate(&Metric{})
	database.ORM.AutoMigrate(&MetricLabel{})
}
