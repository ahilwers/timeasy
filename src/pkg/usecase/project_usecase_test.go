package usecase

import (
	"errors"
	"fmt"
	"testing"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_projectUsecase_AddProject(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	prj := model.Project{
		Name:   "Testproject",
		UserId: userId,
	}
	err := usecaseTest.ProjectUsecase.AddProject(&prj)
	assert.Nil(t, err)

	var projectFromDb model.Project
	if err := test.DB.First(&projectFromDb, prj.ID).Error; err != nil {
		t.Errorf("project could not be retrieved: %s", err)
	}
	assert.Equal(t, prj.Name, projectFromDb.Name)
	assert.Equal(t, userId, projectFromDb.UserId)
	assert.Nil(t, projectFromDb.TeamID)
}

func Test_projectUsecase_AddProjectFailsWithoutUserId(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	prj := model.Project{
		Name: "Testproject",
	}
	err := usecaseTest.ProjectUsecase.AddProject(&prj)
	assert.NotNil(t, err)
}

func Test_projectUsecase_GetProjectById(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	prj := model.Project{
		Name:   "Testproject",
		UserId: userId,
	}
	err := usecaseTest.ProjectUsecase.AddProject(&prj)
	assert.Nil(t, err)

	projectFromDb, err := usecaseTest.ProjectUsecase.GetProjectById(prj.ID)
	assert.Nil(t, err)
	assert.Equal(t, prj.ID, projectFromDb.ID)
	assert.Equal(t, prj.Name, projectFromDb.Name)
}

func Test_projectUsecase_GetProjectByIdFailsIfProjectDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)

	_, err = usecaseTest.ProjectUsecase.GetProjectById(notExistingId)
	assert.NotNil(t, err)
}

func Test_projectUsecase_GetAllProjects(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	addProjects(t, usecaseTest.ProjectUsecase, 3, userId)

	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(projectsFromDb))
	for i, project := range projectsFromDb {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
		assert.Equal(t, userId, project.UserId)
	}
}

func Test_projectUsecase_GetAllProjectsOfUser(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	addProjects(t, usecaseTest.ProjectUsecase, 3, userId)
	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	addProjectsWithStartIndex(t, usecaseTest.ProjectUsecase, 4, 3, otherUserId)

	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjectsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(projectsFromDb))
	for i, project := range projectsFromDb {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
		assert.Equal(t, userId, project.UserId)
	}
	projectsFromDb, err = usecaseTest.ProjectUsecase.GetAllProjectsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(projectsFromDb))
	for i, project := range projectsFromDb {
		assert.Equal(t, fmt.Sprintf("Project %v", i+4), project.Name)
		assert.Equal(t, otherUserId, project.UserId)
	}
}

func Test_projectUsecase_UpdateProject(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project1", userId)
	project.Name = "updatedProject"
	err := usecaseTest.ProjectUsecase.UpdateProject(&project)
	assert.Nil(t, err)

	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, project.ID, projectsFromDb[0].ID)
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
}

func Test_projectUsecase_UpdateProjectFailsIfProjectDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	project := model.Project{
		ID:     projectId,
		Name:   "project",
		UserId: userId,
	}

	project.Name = "updatedProject"
	err = usecaseTest.ProjectUsecase.UpdateProject(&project)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectUsecase_UpdateProjectFailsIfItHasNoUserId(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project1", userId)
	project.Name = "updatedProject"
	project.UserId = uuid.Nil
	err := usecaseTest.ProjectUsecase.UpdateProject(&project)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	// The project data should not have been changed:
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, project.ID, projectsFromDb[0].ID)
	assert.Equal(t, "project1", projectsFromDb[0].Name)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
}

func Test_projectUsecase_DeleteProject(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	projects := addProjects(t, usecaseTest.ProjectUsecase, 3, userId)

	err := usecaseTest.ProjectUsecase.DeleteProject(projects[1].ID)
	assert.Nil(t, err)
	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projectsFromDb))
	assert.Equal(t, "Project 1", projectsFromDb[0].Name)
	assert.Equal(t, "Project 3", projectsFromDb[1].Name)
}

func Test_projectUsecase_DeleteProjectFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = usecaseTest.ProjectUsecase.DeleteProject(notExistingId)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_projectUsecase_CanProjectBeAssignedToATeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "Testproject", userId)
	assert.Nil(t, project.TeamID)

	team := model.Team{
		Name1: "Testteam",
	}

	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	err = usecaseTest.ProjectUsecase.AssignProjectToTeam(&project, &team)
	assert.Nil(t, err)

	projectFromDb, err := usecaseTest.ProjectUsecase.GetProjectById(project.ID)
	assert.Nil(t, err)
	assert.Equal(t, *projectFromDb.TeamID, team.ID)
}

func Test_projectUsecase_AssignProjectToTeamFailsIfProjectDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	project := model.Project{
		Name: "NotExistingProject",
	}

	team := model.Team{
		Name1: "Testteam",
	}

	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	err = usecaseTest.ProjectUsecase.AssignProjectToTeam(&project, &team)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_projectUsecase_AssignProjectToTeamFilesIfTeamDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "Testproject", userId)
	assert.Nil(t, project.TeamID)

	team := model.Team{
		Name1: "Testteam",
	}

	err := usecaseTest.ProjectUsecase.AssignProjectToTeam(&project, &team)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func addProjects(t *testing.T, projectUsecase ProjectUsecase, count int, userId uuid.UUID) []model.Project {
	return addProjectsWithStartIndex(t, projectUsecase, 1, count, userId)
}

func addProjectsWithStartIndex(t *testing.T, projectUsecase ProjectUsecase, startIndex int, count int, userId uuid.UUID) []model.Project {
	var projects []model.Project
	for i := 0; i < count; i++ {
		project := addProject(t, projectUsecase, fmt.Sprintf("Project %v", startIndex+i), userId)
		projects = append(projects, project)
	}
	return projects
}

func addProject(t *testing.T, projectUsecase ProjectUsecase, name string, userId uuid.UUID) model.Project {
	prj := model.Project{
		Name:   name,
		UserId: userId,
	}
	err := projectUsecase.AddProject(&prj)
	assert.Nil(t, err)
	return prj
}
