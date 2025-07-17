package usecase

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"
)

func Test_projectUsecase_AddProject(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	hourlyRate, err := decimal.NewFromString("10.23")
	assert.Nil(t, err)

	prj := model.Project{
		Name:       "Testproject",
		UserId:     userId,
		Deadline:   model.NewDateOnly(time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC)),
		HourlyRate: hourlyRate,
		TimeBudget: 50,
	}
	err = usecaseTest.ProjectUsecase.AddProject(&prj)
	assert.Nil(t, err)

	projectFromDb, err := usecaseTest.ProjectUsecase.GetProjectById(prj.ID)
	if err != nil {
		t.Errorf("project could not be retrieved: %s", err)
	}
	assert.Equal(t, prj.Name, projectFromDb.Name)
	assert.Equal(t, userId, projectFromDb.UserId)
	assert.True(t, prj.Deadline.Equal(projectFromDb.Deadline))
	assert.Equal(t, prj.HourlyRate, projectFromDb.HourlyRate)
	assert.Equal(t, prj.TimeBudget, projectFromDb.TimeBudget)
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

func Test_projectUsecase_GetAllProjectsOfUserAlsoReturnsProjectsOfUsersTeams(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	// Create a user and a team for it:
	userId := GetTestUserId(t)
	teamOfUser := model.Team{
		Name1: "Team1OfUser",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&teamOfUser, userId)
	assert.Nil(t, err)

	// Create another user and a team for it
	otherUserId := GetTestUserId(t)
	teamOfOtherUser := model.Team{
		Name1: "Team1OfOtherUser",
	}
	err = usecaseTest.TeamUsecase.AddTeam(&teamOfOtherUser, otherUserId)
	assert.Nil(t, err)

	// Assign the user to the team of the other user:
	_, err = usecaseTest.TeamUsecase.AddUserToTeam(userId, &teamOfOtherUser, model.RoleList{model.RoleUser})

	// Create a project that belongs to the first user:
	projectOfUser := model.Project{
		UserId: userId,
		Name:   "ProjectOfUser",
	}
	err = usecaseTest.ProjectUsecase.AddProject(&projectOfUser)
	assert.Nil(t, err)

	// Create a project that belongs to the other user:
	projectOfOtherUser := model.Project{
		UserId: otherUserId,
		Name:   "ProjectOfOtherUser",
	}
	err = usecaseTest.ProjectUsecase.AddProject(&projectOfOtherUser)
	assert.Nil(t, err)

	// Create a project that belongs to the team of other user and the first user:
	projectOfOtherUsersTeam := model.Project{
		UserId: otherUserId,
		TeamID: &teamOfOtherUser.ID,
		Name:   "ProjectOfOtherUsersTeam",
	}
	err = usecaseTest.ProjectUsecase.AddProject(&projectOfOtherUsersTeam)
	assert.Nil(t, err)

	projectsFromDb, err := usecaseTest.ProjectUsecase.GetAllProjectsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projectsFromDb))
	for i, project := range projectsFromDb {
		expectedId := uuid.Nil
		switch i {
		case 0:
			expectedId = projectOfOtherUsersTeam.ID
		case 1:
			expectedId = projectOfUser.ID
		}
		assert.Equal(t, expectedId, project.ID)
	}
	projectsFromDb, err = usecaseTest.ProjectUsecase.GetAllProjectsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(projectsFromDb))
	for i, project := range projectsFromDb {
		expectedId := uuid.Nil
		switch i {
		case 0:
			expectedId = projectOfOtherUser.ID
		case 1:
			expectedId = projectOfOtherUsersTeam.ID
		}
		assert.Equal(t, expectedId, project.ID)
	}
}

func Test_projectUsecase_AddProject_AlsoAddsChangelogEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "Testproject", GetTestUserId(t))
	changelogEntries, err := usecaseTest.ChangelogRepo.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changelogEntries))
	changelogEntry := changelogEntries[0]
	assert.Equal(t, model.EntityTypeProject, changelogEntry.EntityType)
	assert.Equal(t, project.ID, changelogEntry.EntityID)
	assert.Equal(t, model.OperationCreated, changelogEntries[0].Operation)
}

func Test_projectUsecase_UpdateProject_AlsoAddsChangelogEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "Testproject", GetTestUserId(t))
	changelogEntries, err := usecaseTest.ChangelogRepo.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changelogEntries))
	assert.Equal(t, model.EntityTypeProject, changelogEntries[0].EntityType)
	assert.Equal(t, project.ID, changelogEntries[0].EntityID)

	usecaseTest.ProjectUsecase.UpdateProject(&project)
	changelogEntries, err = usecaseTest.ChangelogRepo.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(changelogEntries))
	assert.Equal(t, model.EntityTypeProject, changelogEntries[0].EntityType)
	assert.Equal(t, model.OperationCreated, changelogEntries[0].Operation)
	assert.Equal(t, project.ID, changelogEntries[0].EntityID)
	assert.Equal(t, model.EntityTypeProject, changelogEntries[1].EntityType)
	assert.Equal(t, project.ID, changelogEntries[1].EntityID)
	assert.Equal(t, model.OperationUpdated, changelogEntries[1].Operation)
}

func Test_projectUsecase_DeleteProject_AlsoAddsChangelogEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "Testproject", GetTestUserId(t))
	changelogEntries, err := usecaseTest.ChangelogRepo.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changelogEntries))
	assert.Equal(t, model.EntityTypeProject, changelogEntries[0].EntityType)
	assert.Equal(t, project.ID, changelogEntries[0].EntityID)

	usecaseTest.ProjectUsecase.DeleteProject(project.ID)
	changelogEntries, err = usecaseTest.ChangelogRepo.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(changelogEntries))
	assert.Equal(t, model.EntityTypeProject, changelogEntries[0].EntityType)
	assert.Equal(t, model.OperationCreated, changelogEntries[0].Operation)
	assert.Equal(t, project.ID, changelogEntries[0].EntityID)
	assert.Equal(t, model.EntityTypeProject, changelogEntries[1].EntityType)
	assert.Equal(t, project.ID, changelogEntries[1].EntityID)
	assert.Equal(t, model.OperationDeleted, changelogEntries[1].Operation)
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
