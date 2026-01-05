package test

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"testing"
	"timeasy-server/pkg/database/postgresql"
)

var Database postgresql.Database

func SetupDatabase() (*dockertest.Pool, *dockertest.Resource) {
	log.Println("Trying to start database server.")
	pool, err := dockertest.NewPool("")

	if err != nil {
		log.Fatalf("Could not connect to dokcer: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_USER=dbuser",
			"POSTGRES_PASSWORD=dbpassword",
			"POSTGRES_DB=timeasy_test",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
		// Ensure port bindings are set
		config.PortBindings = map[docker.Port][]docker.PortBinding{
			"5432/tcp": {{HostIP: "", HostPort: ""}},
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	hostAndPort := resource.GetHostPort("5432/tcp")
	port := resource.GetPort("5432/tcp")
	log.Printf("Host and Port: %s, Port: %s\n", hostAndPort, port)

	// Replace localhost with 127.0.0.1 to force IPv4 on macOS
	if port == "" {
		log.Fatal("Could not get port from docker resource - port bindings may not be configured correctly")
	}
	hostAndPort = fmt.Sprintf("127.0.0.1:%s", port)
	connectionString := fmt.Sprintf("postgres://dbuser:dbpassword@%s/timeasy_test?sslmode=disable", hostAndPort)
	// retry until db server is ready
	err = pool.Retry(func() error {
		return connectSqlDb(connectionString)
	})
	log.Println("=========================================================")
	err = Database.Migrate()
	if err != nil {
		log.Fatalf("Could not migrate database: %s", err)
	}
	return pool, resource
}

func connectSqlDb(connectionString string) error {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	Database.DB = db
	pingError := db.Ping()
	return pingError
}

func connectGormDb(connectionString string) error {
	DB, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}
	db, err := DB.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}

func TeardownDatabase(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func SetupTest(tb testing.TB) func(tb testing.TB) {
	err := deleteAllEntities(Database.DB)
	if err != nil {
		tb.Errorf(err.Error())
	}
	return func(tb testing.TB) {
		err := deleteAllEntities(Database.DB)
		if err != nil {
			tb.Errorf(err.Error())
		}
	}
}

func deleteAllEntities(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM change_log")
	if err != nil {
		return err
	}
	// Reset the sequence for change_log table
	_, err = db.Exec("ALTER SEQUENCE change_log_id_seq RESTART WITH 1")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM time_entries")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM projects")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM user_team_assignments")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM teams")
	if err != nil {
		return err
	}
	return nil
}
