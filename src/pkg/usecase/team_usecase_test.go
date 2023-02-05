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

	team := model.Team{
		Name: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team)
	assert.Nil(t, err)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "Testteam", teamsFromDb[0].Name)
}

func Test_teamUsecase_UpdateTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	team := model.Team{
		Name: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team)
	assert.Nil(t, err)

	team.Name = "UpdatedTeam"
	err = TestTeamUsecase.UpdateTeam(&team)
	assert.Nil(t, err)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "UpdatedTeam", teamsFromDb[0].Name)
}

func Test_teamUsecase_UpdateTeamFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	team := model.Team{
		Name: "Testteam",
	}

	team.Name = "UpdatedTeam"
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

	team := model.Team{
		Name: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team)
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

	team := model.Team{
		Name: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&team)
	assert.Nil(t, err)

	teamFromDb, err := TestTeamUsecase.GetTeamById(team.ID)
	assert.Nil(t, err)
	assert.Equal(t, team.Name, teamFromDb.Name)
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

	teams := addTeams(t, 3)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(teamsFromDb))
	for i, team := range teamsFromDb {
		assert.Equal(t, teams[i].Name, team.Name)
	}
}

func addTeams(t *testing.T, count int) []model.Team {
	var teams []model.Team
	for i := 0; i < count; i++ {
		team := model.Team{
			Name: fmt.Sprintf("Team %v", i+1),
		}
		teams = append(teams, team)
		err := TestTeamUsecase.AddTeam(&team)
		assert.Nil(t, err)
	}
	return teams
}
