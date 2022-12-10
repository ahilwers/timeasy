package test

import (
	"fmt"
	"log"
	"timeasy-server/pkg/domain/model"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetupDatabase() (*dockertest.Pool, *dockertest.Resource) {
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
		DB, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
		if err != nil {
			return err
		}
		db, err := DB.DB()
		if err != nil {
			return err
		}
		return db.Ping()
	})
	DB.AutoMigrate(&model.Project{})
	return pool, resource
}

func TeardownDatabase(pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}
