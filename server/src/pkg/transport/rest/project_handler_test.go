package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_projectHandler_GetProjectById(t *testing.T) {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	clientId := "1"
	token := authTokenMock{}
	token.On("GetUserId").Return(userId, nil)
	token.On("HasRole", model.RoleUser).Return(true, nil)
	token.On("HasRole", model.RoleAdmin).Return(false, nil)

	verifier := tokenVerifierMock{}
	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)

	handlerTest := NewHandlerTest(&verifier)
	teardownTest := handlerTest.SetupTest(t)
	defer teardownTest(t)

	project := model.Project{
		Name:   "testproject",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, string(clientId))
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, userId, projectFromService.UserId)
}

func Test_projectHandler_GetProjectByIdFailsIfProjectDoesNotExist(t *testing.T) {
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

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", projectId), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", projectId))
}

func Test_projectHandler_GetProjectByIdFailsIfItDoesNotBelongToUser(t *testing.T) {
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

	projectOwnerId, err := uuid.NewV4()
	assert.Nil(t, err)

	project := model.Project{
		Name:   "testproject",
		UserId: projectOwnerId,
	}
	clientId := "1"
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", project.ID))
}

func Test_projectHandler_GetProjectByIdSucceedsIfItBelongsToUsersTeam(t *testing.T) {
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

	project := model.Project{
		Name:   "testproject",
		UserId: otherUserId,
	}

	clientId := "1"
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	team := model.Team{
		Name1: "Team",
	}
	err = handlerTest.TeamUsecase.AddTeam(&team, userId, clientId)
	assert.Nil(t, err)

	err = handlerTest.ProjectUsecase.AssignProjectToTeam(&project, &team, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, otherUserId, projectFromService.UserId)
}

func Test_projectHandler_GetProjectByIdPassesIfBelongsToOtherUserAndUserIsAdmin(t *testing.T) {
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

	projectOwnerId, err := uuid.NewV4()
	assert.Nil(t, err)

	project := model.Project{
		Name:   "testproject",
		UserId: projectOwnerId,
	}
	clientId := "1"
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, projectOwnerId, projectFromService.UserId)
}

func Test_projectHandler_GetAllProjects(t *testing.T) {
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

	addProjects(t, handlerTest, 3, userId)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectsFromService []model.Project
	json.Unmarshal(w.Body.Bytes(), &projectsFromService)
	for i, project := range projectsFromService {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
		assert.Equal(t, userId, project.UserId)
	}
}

func Test_projectHandler_GetAllProjectsReturnsOnlyProjectsOfUser(t *testing.T) {
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

	addProjects(t, handlerTest, 3, userId)

	otherUserId, err := uuid.NewV4()
	addProjectsWithStartIndex(t, handlerTest, 4, 3, otherUserId)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectsFromService []model.Project
	json.Unmarshal(w.Body.Bytes(), &projectsFromService)
	assert.Equal(t, 3, len(projectsFromService))
	for i, project := range projectsFromService {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
		assert.Equal(t, userId, project.UserId)
	}
}

func Test_projectHandler_GetAllProjectsReturnsAllProjectsIfUserIsAdmin(t *testing.T) {
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

	addProjects(t, handlerTest, 3, userId)

	otherUserId, err := uuid.NewV4()
	addProjectsWithStartIndex(t, handlerTest, 4, 3, otherUserId)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectsFromService []model.Project
	json.Unmarshal(w.Body.Bytes(), &projectsFromService)
	assert.Equal(t, 6, len(projectsFromService))
	for i, project := range projectsFromService {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
	}
}

func Test_projectHandler_AddProject(t *testing.T) {
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
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "project1", "#ff0000"))
	req, err := http.NewRequest("POST", "/api/v1/projects", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "project1", projectsFromDb[0].Name)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProject(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", userId)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\", \"deadline\": \"%v\", \"hourlyRate\": %v, \"timeBudget\": %v}", "updatedProject", "#ff0000", "2020-01-01", 23.19, 20))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
	newDeadline := model.NewDateOnly(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	assert.True(t, newDeadline.Equal(projectsFromDb[0].Deadline))
	expectedHourlyRate, err := decimal.NewFromString("23.19")
	assert.Nil(t, err)
	assert.Equal(t, expectedHourlyRate, projectsFromDb[0].HourlyRate)
	assert.Equal(t, 20, projectsFromDb[0].TimeBudget)
}

func Test_projectHandler_UpdateProject_ShouldNotUpdateFieldsThatAreNotProvided(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", userId)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "updatedProject", "#ff0000"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.True(t, project.Deadline.Equal(projectsFromDb[0].Deadline))
	assert.True(t, project.HourlyRate.Equal(projectsFromDb[0].HourlyRate))
	assert.Equal(t, project.TimeBudget, projectsFromDb[0].TimeBudget)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProject_ShouldNotUpdateDeadlineIfItsEmpty(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", userId)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\", \"deadline\": \"%v\"}", "updatedProject", "#ff0000", ""))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.True(t, projectsFromDb[0].Deadline.IsZero())
	assert.True(t, project.HourlyRate.Equal(projectsFromDb[0].HourlyRate))
	assert.Equal(t, project.TimeBudget, projectsFromDb[0].TimeBudget)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProjectAsTeamLead(t *testing.T) {
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
	project := addProject(t, handlerTest, "project", otherUserId)

	team := model.Team{
		Name1: "Team",
	}
	clientId := "1"
	err = handlerTest.TeamUsecase.AddTeam(&team, userId, clientId)
	assert.Nil(t, err)

	err = handlerTest.ProjectUsecase.AssignProjectToTeam(&project, &team, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "updatedProject", "#ff0000"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.Equal(t, otherUserId, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProjectIfUserIsNotTeamLead(t *testing.T) {
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
	project := addProject(t, handlerTest, "project", otherUserId)

	team := model.Team{
		Name1: "Team",
	}
	clientId := "1"
	err = handlerTest.TeamUsecase.AddTeam(&team, otherUserId, clientId)
	assert.Nil(t, err)

	_, err = handlerTest.TeamUsecase.AddUserToTeam(userId, &team, model.RoleList{model.RoleUser}, userId, clientId)
	assert.Nil(t, err)

	err = handlerTest.ProjectUsecase.AssignProjectToTeam(&project, &team, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "updatedProject", "#ff0000"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "project", projectsFromDb[0].Name)
	assert.Equal(t, otherUserId, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProjectFailsIfItDoesNotExist(t *testing.T) {
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

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "updatedProject", "#ff0000"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", projectId), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectHandler_UpdateProjectFailsIfItBelongsToAnotherUser(t *testing.T) {
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

	projectOwnerId, err := uuid.NewV4()
	project := addProject(t, handlerTest, "project", projectOwnerId)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "updatedProject", "#ff0000"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to update this project")

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "project", projectsFromDb[0].Name)
	assert.Equal(t, projectOwnerId, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProjectSucceedsIfItBelongsToAnotherUserAndUserIsAdmin(t *testing.T) {
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

	projectOwnerId, err := uuid.NewV4()
	assert.Nil(t, err)
	project := addProject(t, handlerTest, "project", projectOwnerId)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\", \"color\": \"%v\"}", "updatedProject", "#ff0000"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.Equal(t, projectOwnerId, projectsFromDb[0].UserId)
}

func Test_projectHandler_DeleteProject(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", userId)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectHandler_DeleteProjectAsTeamLead(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", otherUserId)

	team := model.Team{
		Name1: "Team",
	}
	clientId := "1"
	err = handlerTest.TeamUsecase.AddTeam(&team, userId, clientId)
	assert.Nil(t, err)

	err = handlerTest.ProjectUsecase.AssignProjectToTeam(&project, &team, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectHandler_DeleteProjectFailsIfUserIsNotTeamLead(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", otherUserId)

	team := model.Team{
		Name1: "Team",
	}
	clientId := "1"
	err = handlerTest.TeamUsecase.AddTeam(&team, otherUserId, clientId)
	assert.Nil(t, err)

	_, err = handlerTest.TeamUsecase.AddUserToTeam(userId, &team, model.RoleList{model.RoleUser}, userId, clientId)
	assert.Nil(t, err)

	err = handlerTest.ProjectUsecase.AssignProjectToTeam(&project, &team, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
}

func Test_projectHandler_DeleteProjectFailsIfItDoesNotExist(t *testing.T) {
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

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", projectId), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func Test_projectHandler_DeleteProjectFailsIfItBelongsToAnotherUser(t *testing.T) {
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

	projectOwnerId, err := uuid.NewV4()
	assert.Nil(t, err)
	project := addProject(t, handlerTest, "project", projectOwnerId)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
}

func Test_projectHandler_DeleteProjectSucceedsIfItBelongsToAnotherUserAndUserIsAdmin(t *testing.T) {
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

	projectOwnerId, err := uuid.NewV4()
	assert.Nil(t, err)
	project := addProject(t, handlerTest, "project", projectOwnerId)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := handlerTest.ProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectHandler_AssignProjectToTeam(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", userId)

	team := model.Team{
		Name1: "Team",
	}
	clientId := "1"
	err = handlerTest.TeamUsecase.AddTeam(&team, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"projectId\": \"%v\", \"teamId\": \"%v\"}", project.ID, team.ID))
	req, err := http.NewRequest("POST", "/api/v1/projects/team", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	projectFromDb, err := handlerTest.ProjectUsecase.GetProjectById(project.ID)
	assert.Nil(t, err)
	assert.Equal(t, "project", projectFromDb.Name)
	assert.Equal(t, userId, projectFromDb.UserId)
	assert.Equal(t, team.ID, *projectFromDb.TeamID)
}

func Test_projectHandler_AssignProjectToTeamFailsIfUserIsNotTeamAdmin(t *testing.T) {
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

	project := addProject(t, handlerTest, "project", userId)

	teamAdminId, err := uuid.NewV4()
	assert.Nil(t, err)

	team := model.Team{
		Name1: "Team",
	}
	clientId := "1"
	err = handlerTest.TeamUsecase.AddTeam(&team, teamAdminId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"projectId\": \"%v\", \"teamId\": \"%v\"}", project.ID, team.ID))
	req, err := http.NewRequest("POST", "/api/v1/projects/team", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	projectFromDb, err := handlerTest.ProjectUsecase.GetProjectById(project.ID)
	assert.Nil(t, err)
	assert.Equal(t, "project", projectFromDb.Name)
	assert.Equal(t, userId, projectFromDb.UserId)
	assert.Nil(t, projectFromDb.TeamID)
}

func addProjects(t *testing.T, handlerTest *HandlerTest, count int, userId uuid.UUID) []model.Project {
	return addProjectsWithStartIndex(t, handlerTest, 1, count, userId)
}

func addProjectsWithStartIndex(t *testing.T, handlerTest *HandlerTest, startIndex int, count int, userId uuid.UUID) []model.Project {
	var projects []model.Project
	for i := 0; i < count; i++ {
		project := addProject(t, handlerTest, fmt.Sprintf("Project %v", startIndex+i), userId)
		projects = append(projects, project)
	}
	return projects
}

func addProject(t *testing.T, handlerTest *HandlerTest, name string, userId uuid.UUID) model.Project {
	prj := model.Project{
		Name:       name,
		UserId:     userId,
		Color:      "#ff0000",
		HourlyRate: decimal.NewFromInt(20),
		TimeBudget: 10,
		Deadline:   model.NewDateOnly(time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC)),
	}
	clientId := "1"
	err := handlerTest.ProjectUsecase.AddProject(&prj, userId, clientId)
	assert.Nil(t, err)
	return prj
}
