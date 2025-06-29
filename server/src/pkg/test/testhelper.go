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
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	log.Printf("Port: %s\n", resource.GetPort("5432/tcp"))

	connectionString := fmt.Sprintf("host=localhost user=dbuser password=dbpassword dbname=timeasy_test port=%v sslmode=disable", resource.GetPort("5432/tcp"))
	// retry until db server is ready
	err = pool.Retry(func() error {
		/*err = connectGormDb(connectionString)
		  if err != nil {
		  	return err
		  }
		*/
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
	_, err := db.Exec("DELETE FROM time_entries")
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
