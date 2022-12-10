package usecase

import (
	"fmt"
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	log.Println("Testmain")

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
		db, err = gorm.Open(postgres.Open(connectionString), &gorm.Config{})
		if err != nil {
			return err
		}
		db, err := db.DB()
		if err != nil {
			return err
		}
		return db.Ping()
	})
	db.AutoMigrate(&model.Project{})
	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func setupTest(tb testing.TB) func(tb testing.TB) {
	log.Println("setup test")
	deleteAllEntities(db)
	return func(tb testing.TB) {
		log.Println("teardown test")
		deleteAllEntities(db)
	}
}

func deleteAllEntities(db *gorm.DB) error {
	log.Println("Deleting all entities.")
	return db.Exec("DELETE FROM projects").Error
}

func Test_projectService_AddProject(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(db)
	projectUsecase := NewProjectUsecase(projectRepo)
	prj := model.Project{
		Name:   "Testproject",
		UserId: "1",
	}
	_, err := projectUsecase.AddProject(&prj)
	if err != nil {
		t.Errorf("Project could not be created: %s", err)
	}
	var projectFromDb model.Project
	if err := db.First(&projectFromDb, prj.ID).Error; err != nil {
		t.Errorf("project could not be retrieved: %s", err)
	}

	if projectFromDb.Name != prj.Name {
		t.Error("project names are not equal.")
	}
}

func Test_projectService_AddProjectFailsWithoutUserId(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(db)
	projectUsecase := NewProjectUsecase(projectRepo)

	prj := model.Project{
		Name: "Testproject",
	}
	_, err := projectUsecase.AddProject(&prj)
	if err == nil {
		t.Error("adding a project without userid is not allowed.")
	}
}
