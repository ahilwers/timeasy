package usecase

import (
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"
)

func Test_projectService_AddProject(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	if err != nil {
		t.Errorf("user can not be added: %v", err)
	}

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	prj := model.Project{
		Name:   "Testproject",
		UserId: user.ID,
	}
	_, err = projectUsecase.AddProject(&prj)
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
	teardownTest := test.SetupTest(t)
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
