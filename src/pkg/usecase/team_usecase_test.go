package usecase

import (
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/stretchr/testify/assert"
)

func Test_teamUsecase_AddTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	prj := model.Team{
		Name: "Testteam",
	}
	err := TestTeamUsecase.AddTeam(&prj)
	assert.Nil(t, err)

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "Testteam", teamsFromDb[0].Name)
}
