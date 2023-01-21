package usecase

import (
	"fmt"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_projectUsecase_AddProject(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	prj := model.Project{
		Name:   "Testproject",
		UserId: user.ID,
	}
	err := projectUsecase.AddProject(&prj)
	assert.Nil(t, err)

	var projectFromDb model.Project
	if err := test.DB.First(&projectFromDb, prj.ID).Error; err != nil {
		t.Errorf("project could not be retrieved: %s", err)
	}
	assert.Equal(t, prj.Name, projectFromDb.Name)
	assert.Equal(t, user.ID, projectFromDb.UserId)
}

func Test_projectUsecase_AddProjectFailsWithoutUserId(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)

	prj := model.Project{
		Name: "Testproject",
	}
	err := projectUsecase.AddProject(&prj)
	assert.NotNil(t, err)
}

func Test_projectUsecase_GetProjectById(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	prj := model.Project{
		Name:   "Testproject",
		UserId: user.ID,
	}
	err := projectUsecase.AddProject(&prj)
	assert.Nil(t, err)

	projectFromDb, err := projectUsecase.GetProjectById(prj.ID)
	assert.Nil(t, err)
	assert.Equal(t, prj.ID, projectFromDb.ID)
	assert.Equal(t, prj.Name, projectFromDb.Name)
}

func Test_projectUsecase_GetProjectByIdFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)

	_, err = projectUsecase.GetProjectById(notExistingId)
	assert.NotNil(t, err)
}

func Test_projectUsecase_GetAllProjects(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)

	addProjects(t, projectUsecase, 3, user)

	projectsFromDb, err := projectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(projectsFromDb))
	for i, project := range projectsFromDb {
		assert.Equal(t, fmt.Sprintf("Project%v", i+1), project.Name)
		assert.Equal(t, user.ID, project.UserId)
	}
}

func addUser(t *testing.T, username string, password string, roles model.RoleList) model.User {
	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "user",
		Password: "password",
		Roles:    roles,
	}
	_, err := userUsecase.AddUser(&user)
	assert.Nil(t, err)
	return user
}

func addProjects(t *testing.T, projectUsecase ProjectUsecase, count int, user model.User) []model.Project {
	var projects []model.Project
	for i := 0; i < count; i++ {
		project := addProject(t, projectUsecase, fmt.Sprintf("Project%v", i+1), user)
		projects = append(projects, project)
	}
	return projects
}

func addProject(t *testing.T, projectUsecase ProjectUsecase, name string, user model.User) model.Project {
	prj := model.Project{
		Name:   name,
		UserId: user.ID,
	}
	err := projectUsecase.AddProject(&prj)
	assert.Nil(t, err)
	return prj
}
