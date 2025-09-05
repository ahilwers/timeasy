package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"timeasy-server/pkg/database/postgresql"
)

type DatabaseService struct {
	Database postgresql.Database
}

func (databaseService *DatabaseService) Init(host string, databaseName string, user string, password string, port int) error {
	connectionString := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", host, user, password, databaseName, port)
	slog.Info("Opening database connection", "host", host, "port", port, "database", databaseName)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		slog.Error("Failed to open database connection", "error", err)
		return err
	}
	
	// Test the connection
	err = db.Ping()
	if err != nil {
		slog.Error("Failed to ping database", "error", err)
		return err
	}
	slog.Info("Database connection established successfully")
	
	databaseService.Database.DB = db
	err = databaseService.Database.Migrate()
	if err != nil {
		slog.Error("Database migration failed", "error", err)
		return err
	}
	return nil
}
