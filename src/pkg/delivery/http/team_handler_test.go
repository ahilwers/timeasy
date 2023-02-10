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

func Test_teamHandler_GetAllTeamsOnlyReturnsTeamsOfUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_ = addTeams(t, 3, *otherUser)
	userTeams := addTeamsWithStartIndex(t, 3, 4, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamsFromService []testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamsFromService)
	assert.Equal(t, 3, len(teamsFromService))
	for i, teamFromService := range teamsFromService {
		assert.Equal(t, userTeams[i].Name1, teamFromService.Name1)
	}
}

func Test_teamHandler_GetAllTeamsReturnsAllTeamsIfUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	otherTeams := addTeams(t, 3, *otherUser)
	userTeams := addTeamsWithStartIndex(t, 3, 4, user)
	allTeams := append(otherTeams, userTeams[:]...)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamsFromService []testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamsFromService)
	assert.Equal(t, 6, len(teamsFromService))
	for i, teamFromService := range teamsFromService {
		assert.Equal(t, allTeams[i].Name1, teamFromService.Name1)
	}
}

func Test_teamHandler_AddUserToTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	team := addTeam(t, "team", user)

	w := httptest.NewRecorder()

	userToBeAdded, err := addUser("userToBeAdded", "password", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	reader := strings.NewReader(fmt.Sprintf("{\"id\": \"%v\"}", userToBeAdded.ID))
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfUser, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(userToBeAdded.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser}, userToBeAdded.Roles)
}

func Test_teamHandler_AddUserToTeamWithSpecificRole(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	team := addTeam(t, "team", user)

	w := httptest.NewRecorder()

	userToBeAdded, err := addUser("userToBeAdded", "password", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	json := fmt.Sprintf("{\"id\": \"%v\", \"roles\": [\"%v\"]}", userToBeAdded.ID, model.RoleAdmin)
	reader := strings.NewReader(json)
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfUser, err := TestTeamUsecase.GetTeamsOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(userToBeAdded.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleAdmin}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_AddUserToTeamFailsIfUserIsNotAdminOfTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	teamAdminUser, err := addUser("nonadmin", "nonadmin", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", *teamAdminUser)

	assert.Nil(t, err)
	token, nonAdminUser := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	_, err = TestTeamUsecase.AddUserToTeam(&nonAdminUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	userToBeAdded, err := addUser("userToBeAdded", "password", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"id\": \"%v\"}", userToBeAdded.ID))
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to add users to this team")

	teamsOfUSerToBeAdded, err := TestTeamUsecase.GetTeamsOfUser(userToBeAdded.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfUSerToBeAdded))
}

func Test_teamHandler_AddUserToTeamSucceedsIfUserIsGlobalAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	teamAdminUser, err := addUser("nonadmin", "nonadmin", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", *teamAdminUser)

	assert.Nil(t, err)
	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})

	userToBeAdded, err := addUser("userToBeAdded", "password", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"id\": \"%v\"}", userToBeAdded.ID))
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfUserToBeAdded, err := TestTeamUsecase.GetTeamsOfUser(userToBeAdded.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUserToBeAdded))
	assert.Equal(t, team.ID, teamsOfUserToBeAdded[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleUser}, teamsOfUserToBeAdded[0].Roles)
}

func Test_teamHandler_DeleteUserFromTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	team := addTeam(t, "team", user)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_, err = TestTeamUsecase.AddUserToTeam(otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUser.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))
}

func Test_teamHandler_DeleteUserFromTeamFailsIfUserIsNotTeamAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	teamAdminUser, err := addUser("teamAdmin", "teamPassword", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", *teamAdminUser)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_, err = TestTeamUsecase.AddUserToTeam(otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUser.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
}

func Test_teamHandler_DeleteUserFromTeamSucceedsIfUserIsGlobalAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	teamAdminUser, err := addUser("teamAdmin", "teamPassword", model.RoleList{model.RoleUser})
	team := addTeam(t, "team", *teamAdminUser)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_, err = TestTeamUsecase.AddUserToTeam(otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUser.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))
}

func Test_teamHandler_DeleteUserFromTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	team := addTeam(t, "team", user)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUser.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
}

func Test_teamHandler_UpdateUserRolesInTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	team := addTeam(t, "team", user)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_, err = TestTeamUsecase.AddUserToTeam(otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_UpdateUserRolesInTeamFailsIfUserIsNotTeamAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	teamAdmin, err := addUser("teamAdmin", "teamPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	team := addTeam(t, "team", *teamAdmin)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	_, err = TestTeamUsecase.AddUserToTeam(&user, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_, err = TestTeamUsecase.AddUserToTeam(otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to update users in this team")

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, model.RoleList{model.RoleUser}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_UpdateUserRolesInTeamSucceedsIfUserIsGlobalAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	teamAdmin, err := addUser("teamAdmin", "teamPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	team := addTeam(t, "team", *teamAdmin)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})
	_, err = TestTeamUsecase.AddUserToTeam(&user, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	_, err = TestTeamUsecase.AddUserToTeam(otherUser, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_UpdateUserRolesInTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	team := addTeam(t, "team", user)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = TestTeamUsecase.GetTeamsOfUser(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))
}

func addTeams(t *testing.T, count int, owner model.User) []model.Team {
	return addTeamsWithStartIndex(t, count, 1, owner)
}

func addTeamsWithStartIndex(t *testing.T, count int, startIndex int, owner model.User) []model.Team {
	var teams []model.Team
	for i := 0; i < count; i++ {
		team := addTeam(t, fmt.Sprintf("team %v", i+startIndex), owner)
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
