package rest

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
	"github.com/stretchr/testify/mock"
)

func Test_teamHandler_AddTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"name1\": \"%v\"}", "team1"))
	req, err := http.NewRequest("POST", "/api/v1/teams", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := handlerTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "team1", teamsFromDb[0].Name1)
}

func Test_teamHandler_UpdateTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team1", userId)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"name1\": \"%v\"}", "updatedteam"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v", team.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := handlerTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsFromDb))
	assert.Equal(t, "updatedteam", teamsFromDb[0].Name1)
}

func Test_teamHandler_UpdateTeamFailsIfItDoesNotExist(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	w := httptest.NewRecorder()
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	reader := strings.NewReader(fmt.Sprintf("{\"name1\": \"%v\"}", "updatedteam"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v", missingId), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := handlerTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamHandler_DeleteTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team1", userId)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v", team.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := handlerTest.TeamUsecase.GetAllTeams()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsFromDb))
}

func Test_teamHandler_DeleteTeamFailsIfItDoesNotExist(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	_ = addTeam(t, handlerTest, "team1", userId)

	w := httptest.NewRecorder()
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v", missingId), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsFromDb, err := handlerTest.TeamUsecase.GetAllTeams()
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
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team1", userId)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/teams/%v", team.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamFromService testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamFromService)
	assert.Equal(t, team.Name1, teamFromService.Name1)
}

func Test_teamHandler_GetTeamByIdFailsIfItDoesNotExist(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	w := httptest.NewRecorder()
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/teams/%v", missingId), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("team with id %v not found", missingId))
}

func Test_teamHandler_GetAllTeamsOnlyReturnsTeamsOfUser(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	_ = addTeams(t, handlerTest, 3, otherUserId)
	userTeams := addTeamsWithStartIndex(t, handlerTest, 3, 4, userId)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamsFromService []testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamsFromService)
	assert.Equal(t, 3, len(teamsFromService))
	for i, teamFromService := range teamsFromService {
		assert.Equal(t, userTeams[i].Name1, teamFromService.Name1)
	}
}

func Test_teamHandler_GetAllTeamsReturnsAllTeamsIfUserIsAdmin(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(true, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	otherTeams := addTeams(t, handlerTest, 3, otherUserId)
	userTeams := addTeamsWithStartIndex(t, handlerTest, 3, 4, userId)
	allTeams := append(otherTeams, userTeams[:]...)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/teams", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var teamsFromService []testTeamDto
	json.Unmarshal(w.Body.Bytes(), &teamsFromService)
	assert.Equal(t, 6, len(teamsFromService))
	for i, teamFromService := range teamsFromService {
		assert.Equal(t, allTeams[i].Name1, teamFromService.Name1)
	}
}

func Test_teamHandler_AddUserToTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team", userId)

	w := httptest.NewRecorder()

	userToBeAddedId, err := uuid.NewV4()
	assert.Nil(t, err)
	reader := strings.NewReader(fmt.Sprintf("{\"id\": \"%v\"}", userToBeAddedId))
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(userToBeAddedId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)
}

func Test_teamHandler_AddUserToTeamWithSpecificRole(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team", userId)

	w := httptest.NewRecorder()

	userToBeAddedId, err := uuid.NewV4()
	assert.Nil(t, err)
	json := fmt.Sprintf("{\"id\": \"%v\", \"roles\": [\"%v\"]}", userToBeAddedId, model.RoleAdmin)
	reader := strings.NewReader(json)
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfUser))
	assert.Equal(t, team.ID, teamsOfUser[0].TeamID)

	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(userToBeAddedId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)
	assert.Equal(t, model.RoleList{model.RoleAdmin}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_AddUserToTeamFailsIfUserIsNotAdminOfTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	// Define an admin for the team:
	teamAdminUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	team := addTeam(t, handlerTest, "team", teamAdminUserId)

	// Add the logged in user to the team
	_, err = handlerTest.TeamUsecase.AddUserToTeam(userId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	userToBeAddedId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	// Try to add another user being the logged in user who is not an admin:
	reader := strings.NewReader(fmt.Sprintf("{\"id\": \"%v\"}", userToBeAddedId))
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/v1/teams/%v/users", team.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to add users to this team")

	teamsOfUSerToBeAdded, err := handlerTest.TeamUsecase.GetTeamsOfUser(userToBeAddedId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfUSerToBeAdded))
}

func Test_teamHandler_DeleteUserFromTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team", userId)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = handlerTest.TeamUsecase.AddUserToTeam(otherUserId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUserId), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))
}

func Test_teamHandler_DeleteUserFromTeamFailsIfUserIsNotTeamAdmin(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	teamAdminUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	team := addTeam(t, handlerTest, "team", teamAdminUserId)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = handlerTest.TeamUsecase.AddUserToTeam(otherUserId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUserId), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
}

func Test_teamHandler_DeleteUserFromTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team", userId)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))

	w := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/teams/%v/users/%v", team.ID, otherUserId), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
}

func Test_teamHandler_UpdateUserRolesInTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team", userId)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = handlerTest.TeamUsecase.AddUserToTeam(otherUserId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUserId), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, model.RoleList{model.RoleUser, model.RoleAdmin}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_UpdateUserRolesInTeamFailsIfUserIsNotTeamAdmin(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	teamAdminId, err := uuid.NewV4()
	assert.Nil(t, err)
	team := addTeam(t, handlerTest, "team", teamAdminId)

	_, err = handlerTest.TeamUsecase.AddUserToTeam(userId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = handlerTest.TeamUsecase.AddUserToTeam(otherUserId, &team, model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, team.ID, teamsOfOtherUser[0].TeamID)

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUserId), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to update users in this team")

	teamsOfOtherUser, err = handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(teamsOfOtherUser))
	assert.Equal(t, model.RoleList{model.RoleUser}, teamsOfOtherUser[0].Roles)
}

func Test_teamHandler_UpdateUserRolesInTeamFailsIfUserDoesNotBelongToTeam(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	team := addTeam(t, handlerTest, "team", userId)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	teamsOfOtherUser, err := handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))

	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"roles\": [\"%v\", \"%v\"]}", model.RoleUser, model.RoleAdmin))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/teams/%v/users/%v/roles", team.ID, otherUserId), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	teamsOfOtherUser, err = handlerTest.TeamUsecase.GetTeamsOfUser(otherUserId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(teamsOfOtherUser))
}

func addTeams(t *testing.T, handlerTest *HandlerTest, count int, ownerId uuid.UUID) []model.Team {
	return addTeamsWithStartIndex(t, handlerTest, count, 1, ownerId)
}

func addTeamsWithStartIndex(t *testing.T, handlerTest *HandlerTest, count int, startIndex int, ownerId uuid.UUID) []model.Team {
	var teams []model.Team
	for i := 0; i < count; i++ {
		team := addTeam(t, handlerTest, fmt.Sprintf("team %v", i+startIndex), ownerId)
		teams = append(teams, team)
	}
	return teams
}

func addTeam(t *testing.T, handlerTest *HandlerTest, name string, ownerId uuid.UUID) model.Team {
	team := model.Team{
		Name1: name,
	}
	err := handlerTest.TeamUsecase.AddTeam(&team, ownerId)
	assert.Nil(t, err)
	return team
}
