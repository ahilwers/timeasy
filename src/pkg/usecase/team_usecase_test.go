package usecase

import (
	"errors"
	"fmt"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_teamUsecase_AddTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	team := model.Team{
		Name1: "Testteam",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	teamsFromDb, err := usecaseTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "Testteam", teamsFromDb[0].Name1)

	assignments, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignments))
	assert.Equal(t, userId, assignments[0].UserID)
	assert.Equal(t, team.ID, assignments[0].TeamID)
}

func Test_teamUsecase_UpdateTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := model.Team{
		Name1: "Testteam",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	team.Name1 = "UpdatedTeam"
	err = usecaseTest.TeamUsecase.UpdateTeam(&team)
	assert.Nil(t, err)

	teamsFromDb, err := usecaseTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "UpdatedTeam", teamsFromDb[0].Name1)
}

func Test_teamUsecase_UpdateTeamFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	team := model.Team{
		Name1: "Testteam",
	}

	team.Name1 = "UpdatedTeam"
	err := usecaseTest.TeamUsecase.UpdateTeam(&team)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	teamsFromDb, err := usecaseTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamUsecase_DeleteTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := model.Team{
		Name1: "Testteam",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	err = usecaseTest.TeamUsecase.DeleteTeam(team.ID)
	assert.Nil(t, err)

	teamsFromDb, err := usecaseTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamUsecase_DeleteTeamFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = usecaseTest.TeamUsecase.DeleteTeam(missingId)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_teamUsecase_GetTeamById(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := model.Team{
		Name1: "Testteam",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	teamFromDb, err := usecaseTest.TeamUsecase.GetTeamById(team.ID)
	assert.Nil(t, err)
	assert.Equal(t, team.Name1, teamFromDb.Name1)
}

func Test_teamUsecase_GetTeamByIdFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = usecaseTest.TeamUsecase.GetTeamById(missingId)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_teamUsecase_GetAllTeams(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	teams := addTeams(t, usecaseTest.TeamUsecase, 3, userId)

	teamsFromDb, err := usecaseTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(teamsFromDb))
	for i, team := range teamsFromDb {
		assert.Equal(t, teams[i].Name1, team.Name1)
	}
}

func Test_teamUsecase_AddUserToTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := model.Team{
		Name1: "Testteam",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	otherUserId := GetTestUserId(t)
	assignment, err := usecaseTest.TeamUsecase.AddUserToTeam(otherUserId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	assert.Equal(t, otherUserId, assignment.UserID)
	assert.Equal(t, team.ID, assignment.TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser}, assignment.Roles)

	assignmentsFromDb, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignmentsFromDb))
	assert.Equal(t, userId, assignmentsFromDb[0].UserID)
	assert.Equal(t, team.ID, assignmentsFromDb[0].TeamID)
	assert.Equal(t, team.Name1, assignmentsFromDb[0].Team.Name1)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, assignmentsFromDb[0].Roles)

	assignmentsFromDb, err = usecaseTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignmentsFromDb))
	assert.Equal(t, otherUserId, assignmentsFromDb[0].UserID)
	assert.Equal(t, team.ID, assignmentsFromDb[0].TeamID)
	assert.Equal(t, team.Name1, assignmentsFromDb[0].Team.Name1)
	assert.Equal(t, model.RoleList{model.RoleUser}, assignmentsFromDb[0].Roles)
}

func Test_teamUsecase_AddUserToTeamFailsIfAssignmentAlreadyExists(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := model.Team{
		Name1: "Testteam",
	}
	err := usecaseTest.TeamUsecase.AddTeam(&team, userId)
	assert.Nil(t, err)

	_, err = usecaseTest.TeamUsecase.AddUserToTeam(userId, &team, model.RoleList{model.RoleAdmin, model.RoleUser})
	assert.NotNil(t, err)
	var entityExistsError *EntityExistsError
	assert.True(t, errors.As(err, &entityExistsError))
}

func Test_teamUsecase_GetTeamsOfUser(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	// The user is automatically assigned when adding the teams:
	teams := addTeams(t, usecaseTest.TeamUsecase, 3, userId)

	assignments, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(assignments))

	for i, assignment := range assignments {
		assert.Equal(t, userId, assignment.UserID)
		assert.Equal(t, teams[i].ID, assignment.TeamID)
		assert.Equal(t, teams[i].Name1, assignment.Team.Name1)
		assert.Equal(t, teams[i].Name2, assignment.Team.Name2)
		assert.Equal(t, teams[i].Name3, assignment.Team.Name3)
		assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, assignment.Roles)
	}
}

func Test_teamUsecase_DeleteUserFromTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := addTeam(t, usecaseTest.TeamUsecase, "team", userId)
	teamsOfUser, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	err = usecaseTest.TeamUsecase.DeleteUserFromTeam(userId, &team)
	assert.Nil(t, err)
	teamsOfUser, err = usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfUser))
}

func Test_teamUsecase_DeleteUserFromTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := addTeam(t, usecaseTest.TeamUsecase, "team", userId)
	teamsOfUser, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	otherUserId := GetTestUserId(t)

	err = usecaseTest.TeamUsecase.DeleteUserFromTeam(otherUserId, &team)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	// Check if the original team assignment still exists
	teamsOfUser, err = usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
}

func Test_teamUsecase_UpdateUserRolesInTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := addTeam(t, usecaseTest.TeamUsecase, "team", userId)
	teamsOfUser, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfUser[0].Roles)

	err = usecaseTest.TeamUsecase.UpdateUserRolesInTeam(userId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfUser, err = usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser}, teamsOfUser[0].Roles)
}

func Test_teamUsecase_UpdateUserRolesInTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	team := addTeam(t, usecaseTest.TeamUsecase, "team", userId)
	teamsOfUser, err := usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfUser[0].Roles)

	otherUserId := GetTestUserId(t)
	err = usecaseTest.TeamUsecase.UpdateUserRolesInTeam(otherUserId, &team, model.RoleList{model.RoleUser})
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	// Check if the user has still the same roles in the team:
	teamsOfUser, err = usecaseTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfUser[0].Roles)
}

func addTeams(t *testing.T, teamUsecase TeamUsecase, count int, ownerId uuid.UUID) []model.Team {
	var teams []model.Team
	for i := 0; i < count; i++ {
		team := addTeam(t, teamUsecase, fmt.Sprintf("Team %v", i+1), ownerId)
		teams = append(teams, team)
	}
	return teams
}

func addTeam(t *testing.T, teamUsecase TeamUsecase, name string, ownerId uuid.UUID) model.Team {
	team := model.Team{
		Name1: fmt.Sprintf("%v.1", name),
		Name2: fmt.Sprintf("%v.2", name),
		Name3: fmt.Sprintf("%v.3", name),
	}
	err := teamUsecase.AddTeam(&team, ownerId)
	assert.Nil(t, err)
	return team
}
