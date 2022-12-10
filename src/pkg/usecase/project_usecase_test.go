package usecase

import (
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}

func setupTest(tb testing.TB) func(tb testing.TB) {
	log.Println("setup test")
	deleteAllEntities(test.DB)
	return func(tb testing.TB) {
		log.Println("teardown test")
		deleteAllEntities(test.DB)
	}
}

func deleteAllEntities(db *gorm.DB) error {
	log.Println("Deleting all entities.")
	return db.Exec("DELETE FROM projects").Error
}

func Test_projectService_AddProject(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
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
	if err := test.DB.First(&projectFromDb, prj.ID).Error; err != nil {
		t.Errorf("project could not be retrieved: %s", err)
	}

	if projectFromDb.Name != prj.Name {
		t.Error("project names are not equal.")
	}
}

func Test_projectService_AddProjectFailsWithoutUserId(t *testing.T) {
	teardownTest := setupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)

	prj := model.Project{
		Name: "Testproject",
	}
	_, err := projectUsecase.AddProject(&prj)
	if err == nil {
		t.Error("adding a project without userid is not allowed.")
	}
}
