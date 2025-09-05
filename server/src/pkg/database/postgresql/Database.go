package postgresql

import (
	"database/sql"
	"embed"
	"errors"
	"log/slog"
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
	slog.Info("Starting database migrations")
	
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		slog.Error("Failed to create postgres driver", "error", err)
		return err
	}
	
	srcDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		slog.Error("Failed to create migration source", "error", err)
		return err
	}
	
	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", driver)
	if err != nil {
		slog.Error("Failed to create migrate instance", "error", err)
		return err
	}
	
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("Migration failed", "error", err)
		return err
	}
	
	if errors.Is(err, migrate.ErrNoChange) {
		slog.Info("No new migrations to apply")
	} else {
		slog.Info("Migrations applied successfully")
	}
	
	return nil
}
