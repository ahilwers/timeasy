package postgresql

import (
	"database/sql"
	"embed"
	"errors"
	"log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Database struct {
	DB *sql.DB
}

func (db *Database) Migrate() error {
	log.Println("Starting database migrations...")
	
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Printf("Failed to create postgres driver: %v", err)
		return err
	}
	
	srcDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		log.Printf("Failed to create migration source: %v", err)
		return err
	}
	
	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", driver)
	if err != nil {
		log.Printf("Failed to create migrate instance: %v", err)
		return err
	}
	
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migration failed: %v", err)
		return err
	}
	
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No new migrations to apply")
	} else {
		log.Println("Migrations applied successfully")
	}
	
	return nil
}
