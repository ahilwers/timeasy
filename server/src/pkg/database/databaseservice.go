package database

import (
	"database/sql"
	"fmt"
	"log"
	"timeasy-server/pkg/database/postgresql"
)

type DatabaseService struct {
	Database postgresql.Database
}

func (databaseService *DatabaseService) Init(host string, databaseName string, user string, password string, port int) error {
	connectionString := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", host, user, password, databaseName, port)
	log.Printf("Opening database connection...")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Failed to open database connection: %v", err)
		return err
	}
	
	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		return err
	}
	log.Printf("Database connection established successfully")
	
	databaseService.Database.DB = db
	err = databaseService.Database.Migrate()
	if err != nil {
		log.Printf("Database migration failed: %v", err)
		return err
	}
	return nil
}
