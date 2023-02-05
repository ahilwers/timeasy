package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_teamHandler_AddTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"name1\": \"%v\"}", "team1"))
	req, err := http.NewRequest("POST", "/api/v1/teams", reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "team1", teamsFromDb[0].Name1)
}

func Test_teamHandler_UpdateTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	team := addTeam(t, "team1", user)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"name1\": \"%v\"}", "updatedteam"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v", team.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "updatedteam", teamsFromDb[0].Name1)
}

func Test_teamHandler_UpdateTeamFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	w := httptest.NewRecorder()
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	reader := strings.NewReader(fmt.Sprintf("{\"name1\": \"%v\"}", "updatedteam"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v", missingId), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamHandler_DeleteTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	team := addTeam(t, "team1", user)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v", team.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamHandler_DeleteTeamFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	_ = addTeam(t, "team1", user)

	w := httptest.NewRecorder()
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v", missingId), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := TestTeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
}

type testTeamDto struct {
	ID uuid.UUID
	teamInputDto
	Name1 string `json:"name1" binding:"required"`
	Name2 string `json:"name2"`
	Name3 string `json:"name3"`
}

func Test_teamHandler_GetTeamById(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	team := addTeam(t, "team1", user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/teams/%v", team.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamFromService testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamFromService)
	assert.Equal(t, team.Name1, teamFromService.Name1)
}

func Test_teamHandler_GetTeamByIdFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	w := httptest.NewRecorder()
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/teams/%v", missingId), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("team with id %v not found", missingId))
}

func Test_teamHandler_GetAllTeams(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	teams := addTeams(t, 3, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamsFromService []testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamsFromService)
	for i, teamFromService := range teamsFromService {
		assert.Equal(t, teams[i].Name1, teamFromService.Name1)
	}
}

func addTeams(t *testing.T, count int, owner model.User) []model.Team {
	var teams []model.Team
	for i := 0; i < count; i++ {
		team := addTeam(t, fmt.Sprintf("team %v", i+1), owner)
		teams = append(teams, team)
	}
	return teams
}

func addTeam(t *testing.T, name string, owner model.User) model.Team {
	team := model.Team{
		Name1: name,
	}
	err := TestTeamUsecase.AddTeam(&team, &owner)
	assert.Nil(t, err)
	return team
}
