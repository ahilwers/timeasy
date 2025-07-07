package database

import (
	"database/sql"
	"fmt"
	"timeasy-server/pkg/database/postgresql"
)

type DatabaseService struct {
	Database postgresql.Database
}

func (databaseService *DatabaseService) Init(host string, databaseName string, user string, password string, port int) error {
	connectionString := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", host, user, password, databaseName, port)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	databaseService.Database.DB = db
	databaseService.Database.Migrate()
	return nil
}
