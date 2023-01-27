package test

import (
	"fmt"
	"log"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

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

	connectionString := fmt.Sprintf("host=localhost user=dbuser password=dbpassword dbname=timeasy_test port=%v", resource.GetPort("5432/tcp"))
	// retry until db server is ready
	err = pool.Retry(func() error {
		DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{
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
	})
	log.Println("=========================================================")
	DB.AutoMigrate(&model.User{})
	DB.AutoMigrate(&model.Project{})
	DB.AutoMigrate(&model.TimeEntry{})
	return pool, resource
}

func TeardownDatabase(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func SetupTest(tb testing.TB) func(tb testing.TB) {
	err := deleteAllEntities(DB)
	if err != nil {
		tb.Errorf(err.Error())
	}
	return func(tb testing.TB) {
		err := deleteAllEntities(DB)
		if err != nil {
			tb.Errorf(err.Error())
		}
	}
}

func deleteAllEntities(db *gorm.DB) error {
	err := db.Exec("DELETE FROM users")
	if err.Error != nil {
		return err.Error
	}
	err = db.Exec("DELETE FROM projects")
	if err.Error != nil {
		return err.Error
	}
	return nil
}
