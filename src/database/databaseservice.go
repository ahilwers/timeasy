package database

import (
	"fmt"
	"timeasy-server/project"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseService struct {
	Database *gorm.DB
}

func (databaseService *DatabaseService) Init(host string, databaseName string, user string, password string, port int) error {
	connectionString := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v", host, user, password, databaseName, port)
	database, databaseError := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if databaseError != nil {
		return databaseError
	}
	database.AutoMigrate(&project.Project{})

	databaseService.Database = database
	return nil
}
