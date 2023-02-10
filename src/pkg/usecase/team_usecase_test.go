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
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	team := model.Team{
		Name1: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team, &user)
	assert.Nil(t, err)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "Testteam", teamsFromDb[0].Name1)

	assignments, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignments))
	assert.Equal(t, user.ID, assignments[0].UserID)
	assert.Equal(t, team.ID, assignments[0].TeamID)
}

func Test_teamUsecase_UpdateTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := model.Team{
		Name1: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team, &user)
	assert.Nil(t, err)

	team.Name1 = "UpdatedTeam"
	err = TestTeamUsecase.UpdateTeam(&team)
	assert.Nil(t, err)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "UpdatedTeam", teamsFromDb[0].Name1)
}

func Test_teamUsecase_UpdateTeamFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	team := model.Team{
		Name1: "Testteam",
	}

	team.Name1 = "UpdatedTeam"
	err := TestTeamUsecase.UpdateTeam(&team)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamUsecase_DeleteTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := model.Team{
		Name1: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team, &user)
	assert.Nil(t, err)

	err = TestTeamUsecase.DeleteTeam(team.ID)
	assert.Nil(t, err)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamUsecase_DeleteTeamFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = TestTeamUsecase.DeleteTeam(missingId)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_teamUsecase_GetTeamById(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := model.Team{
		Name1: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team, &user)
	assert.Nil(t, err)

	teamFromDb, err := TestTeamUsecase.GetTeamById(team.ID)
	assert.Nil(t, err)
	assert.Equal(t, team.Name1, teamFromDb.Name1)
}

func Test_teamUsecase_GetTeamByIdFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = TestTeamUsecase.GetTeamById(missingId)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_teamUsecase_GetAllTeams(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	teams := addTeams(t, 3, user)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(teamsFromDb))
	for i, team := range teamsFromDb {
		assert.Equal(t, teams[i].Name1, team.Name1)
	}
}

func Test_teamUsecase_AddUserToTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := model.Team{
		Name1: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team, &user)
	assert.Nil(t, err)

	otherUser := addUser(t, "otherUser", "password", model.RoleList{model.RoleUser})
	assignment, err := TestTeamUsecase.AddUserToTeam(&otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	assert.Equal(t, otherUser.ID, assignment.UserID)
	assert.Equal(t, team.ID, assignment.TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser}, assignment.Roles)

	assignmentsFromDb, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignmentsFromDb))
	assert.Equal(t, user.ID, assignmentsFromDb[0].UserID)
	assert.Equal(t, team.ID, assignmentsFromDb[0].TeamID)
	assert.Equal(t, team.Name1, assignmentsFromDb[0].Team.Name1)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, assignmentsFromDb[0].Roles)

	assignmentsFromDb, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(assignmentsFromDb))
	assert.Equal(t, otherUser.ID, assignmentsFromDb[0].UserID)
	assert.Equal(t, team.ID, assignmentsFromDb[0].TeamID)
	assert.Equal(t, team.Name1, assignmentsFromDb[0].Team.Name1)
	assert.Equal(t, model.RoleList{model.RoleUser}, assignmentsFromDb[0].Roles)
}

func Test_teamUsecase_AddUserToTeamFailsIfAssignmentAlreadyExists(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := model.Team{
		Name1: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team, &user)
	assert.Nil(t, err)

	_, err = TestTeamUsecase.AddUserToTeam(&user, &team, model.RoleList{model.RoleAdmin, model.RoleUser})
	assert.NotNil(t, err)
	var entityExistsError *EntityExistsError
	assert.True(t, errors.As(err, &entityExistsError))
}

func Test_teamUsecase_GetTeamsOfUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	// The user is automatically assigned when adding the teams:
	teams := addTeams(t, 3, user)

	assignments, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(assignments))

	for i, assignment := range assignments {
		assert.Equal(t, user.ID, assignment.UserID)
		assert.Equal(t, teams[i].ID, assignment.TeamID)
		assert.Equal(t, teams[i].Name1, assignment.Team.Name1)
		assert.Equal(t, teams[i].Name2, assignment.Team.Name2)
		assert.Equal(t, teams[i].Name3, assignment.Team.Name3)
		assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, assignment.Roles)
	}
}

func Test_teamUsecase_DeleteUserFromTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", user)
	teamsOfUser, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	err = TestTeamUsecase.DeleteUserFromTeam(&user, &team)
	assert.Nil(t, err)
	teamsOfUser, err = TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfUser))
}

func Test_teamUsecase_DeleteUserFromTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", user)
	teamsOfUser, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	otherUser := addUser(t, "otherUser", "otherPassword", model.RoleList{model.RoleUser})

	err = TestTeamUsecase.DeleteUserFromTeam(&otherUser, &team)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	// Check if the original team assignment still exists
	teamsOfUser, err = TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
}

func Test_teamUsecase_UpdateUserRolesInTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", user)
	teamsOfUser, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfUser[0].Roles)

	err = TestTeamUsecase.UpdateUserRolesInTeam(&user, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfUser, err = TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser}, teamsOfUser[0].Roles)
}

func Test_teamUsecase_UpdateUserRolesInTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", user)
	teamsOfUser, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfUser[0].Roles)

	otherUser := addUser(t, "otherUser", "otherPassword", model.RoleList{model.RoleUser})
	err = TestTeamUsecase.UpdateUserRolesInTeam(&otherUser, &team, model.RoleList{model.RoleUser})
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))

	// Check if the user has still the same roles in the team:
	teamsOfUser, err = TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfUser[0].Roles)
}

func addTeams(t *testing.T, count int, owner model.User) []model.Team {
	var teams []model.Team
	for i := 0; i < count; i++ {
		team := addTeam(t, fmt.Sprintf("Team %v", i+1), owner)
		teams = append(teams, team)
	}
	return teams
}

func addTeam(t *testing.T, name string, owner model.User) model.Team {
	team := model.Team{
		Name1: fmt.Sprintf("%v.1", name),
		Name2: fmt.Sprintf("%v.2", name),
		Name3: fmt.Sprintf("%v.3", name),
	}
	err := TestTeamUsecase.AddTeam(&team, &owner)
	assert.Nil(t, err)
	return team
}
