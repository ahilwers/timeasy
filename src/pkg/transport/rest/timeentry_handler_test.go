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

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_timeEntryHandler_AddTimeEntry(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"entry1", startTime.Unix(), project.ID))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	projectsFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "entry1", projectsFromDb[0].Description)
	assert.Equal(t, userId, projectsFromDb[0].UserId)
	assert.Equal(t, startTime, projectsFromDb[0].StartTime)
	assert.True(t, projectsFromDb[0].EndTime.IsZero())
}

func Test_timeEntryHandler_AddTimeEntryFailsIfProjectIdMissing(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v}", "entry1", startTime.Unix()))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	projectsFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_timeEntryHandler_AddTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
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

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	missingProjectId, err := uuid.NewV4()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"entry1", startTime.Unix(), missingProjectId))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", missingProjectId))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entriesFromDb))
}

func Test_timeEntryHandler_UpdateTimeEntry(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), timeEntry.ProjectId))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "updatedentry", entriesFromDb[0].Description)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
	assert.Equal(t, timeEntry.ProjectId, entriesFromDb[0].ProjectId)
	assert.Equal(t, timeEntry.UserId, entriesFromDb[0].UserId)
}

func Test_timeEntryHandler_UpdateTimeEntryFailsIfItDoesNotExist(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), project.ID))
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", missingId), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("entry with id %v not found", missingId))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entriesFromDb))
}

func Test_timeEntryHandler_UpdateTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	missingProjectId, err := uuid.NewV4()
	assert.Nil(t, err)
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), missingProjectId))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", missingProjectId))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "timeentry", entriesFromDb[0].Description)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
	assert.Equal(t, timeEntry.ProjectId, entriesFromDb[0].ProjectId)
	assert.Equal(t, timeEntry.UserId, entriesFromDb[0].UserId)
}

func Test_timeEntryHandler_UpdateTimeEntryFailsIfItDoesNotBelongToTheUser(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	ownerId, err := uuid.NewV4()
	assert.Nil(t, err)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      ownerId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), timeEntry.ProjectId))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to update this entry")

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(ownerId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "timeentry", entriesFromDb[0].Description)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
	assert.Equal(t, timeEntry.ProjectId, entriesFromDb[0].ProjectId)
	assert.Equal(t, timeEntry.UserId, entriesFromDb[0].UserId)
}

func Test_timeEntryHandler_UpdateTimeEntrySucceedsIfItDoesNotBelongToTheUserButTheUserIsAdmin(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	ownerId, err := uuid.NewV4()
	assert.Nil(t, err)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      ownerId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), timeEntry.ProjectId))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(ownerId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "updatedentry", entriesFromDb[0].Description)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
	assert.Equal(t, timeEntry.ProjectId, entriesFromDb[0].ProjectId)
	assert.Equal(t, timeEntry.UserId, entriesFromDb[0].UserId)
}

func Test_timeEntryHandler_DeleteTimeEntry(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entriesFromDb))
}

func Test_timeEntryHandler_DeleteTimeEntryFailsIfitDoesNotExist(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	missingId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/timeentries/%v", missingId), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("entry with id %v not found", missingId))
}

func Test_timeEntryHandler_DeleteTimeEntryFailsIfItDoesNotBelongToTheUser(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	ownerId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      ownerId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("entry with id %v not found", timeEntry.ID))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(ownerId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
}

func Test_timeEntryHandler_DeleteTimeEntrySucceedsIfItDoesNotBelongToTheUserButUserIsAdmin(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	ownerId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      ownerId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), nil)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(ownerId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entriesFromDb))
}

type timeEntryTestDto struct {
	Id               uuid.UUID
	Description      string `json:"description" binding:"required"`
	StartTimeUTCUnix int64  `json:"startTimeUTCUnix" binding:"required"`
	EndTimeUTCUnix   int64
	ProjectId        uuid.UUID `json:"projectId" binding:"required"`
}

func Test_timeEntryHandler_GetTimeEntryById(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entryFromService timeEntryTestDto
	json.Unmarshal(w.Body.Bytes(), &entryFromService)
	assert.Equal(t, timeEntry.Description, entryFromService.Description)
	assert.Equal(t, startTime.Unix(), entryFromService.StartTimeUTCUnix)
	assert.Equal(t, int64(0), entryFromService.EndTimeUTCUnix)
	assert.Equal(t, project.ID, entryFromService.ProjectId)
}

func Test_timeEntryHandler_GetTimeEntryByIdFailsIfItDoesNotExist(t *testing.T) {
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

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/timeentries/%v", missingId), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("entry with id %v not found", missingId))
}

func Test_timeEntryHandler_GetTimeEntryByIdFailsIfItDoesNotBelongToTheUser(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	ownerId, err := uuid.NewV4()
	assert.Nil(t, err)
	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      ownerId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("entry with id %v not found", timeEntry.ID))
}

func Test_timeEntryHandler_GetTimeEntryByIdSucceedsIfItDoesNotBelongToTheUserButUserIsAdmin(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	ownerId, err := uuid.NewV4()
	assert.Nil(t, err)
	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      ownerId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entryFromService timeEntryTestDto
	json.Unmarshal(w.Body.Bytes(), &entryFromService)
	assert.Equal(t, timeEntry.Description, entryFromService.Description)
	assert.Equal(t, startTime.Unix(), entryFromService.StartTimeUTCUnix)
	assert.Equal(t, int64(0), entryFromService.EndTimeUTCUnix)
	assert.Equal(t, project.ID, entryFromService.ProjectId)
}

func Test_timeEntryHandler_GetAllTimeEntries(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	addTimeEntries(t, handlerTest, 3, userId, project)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/timeentries", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entriesFromService []timeEntryDto
	json.Unmarshal(w.Body.Bytes(), &entriesFromService)
	assert.Equal(t, 3, len(entriesFromService))
	for index, entryFromService := range entriesFromService {
		assert.Equal(t, fmt.Sprintf("entry %v", index+1), entryFromService.Description)
		assert.Equal(t, project.ID, entryFromService.ProjectId)
	}
}

func Test_timeEntryHandler_GetAllTimeEntriesOnlyReturnsEntriesOfUser(t *testing.T) {
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	addTimeEntries(t, handlerTest, 3, userId, project)

	otherUserId, err := uuid.NewV4()
	assert.Nil(t, err)
	addTimeEntriesWithStartIndex(t, handlerTest, 4, 3, otherUserId, project)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/timeentries", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entriesFromService []timeEntryDto
	json.Unmarshal(w.Body.Bytes(), &entriesFromService)
	assert.Equal(t, 3, len(entriesFromService))
	for index, entryFromService := range entriesFromService {
		assert.Equal(t, fmt.Sprintf("entry %v", index+1), entryFromService.Description)
		assert.Equal(t, project.ID, entryFromService.ProjectId)
	}
}

func addTimeEntries(t *testing.T, handlerTest *HandlerTest, count int, ownerId uuid.UUID, project model.Project) []model.TimeEntry {
	return addTimeEntriesWithStartIndex(t, handlerTest, 1, count, ownerId, project)
}

func addTimeEntriesWithStartIndex(t *testing.T, handlerTest *HandlerTest, startIndex int, count int, ownerId uuid.UUID, project model.Project) []model.TimeEntry {
	var entries []model.TimeEntry
	startTime := time.Now()
	oneHour := 1000 * 1000 * 60 * 60 // duration is in nanoseconds
	for i := 0; i < count; i++ {
		entry := model.TimeEntry{
			Description: fmt.Sprintf("entry %v", startIndex+i),
			StartTime:   startTime.Add(time.Duration(oneHour * (count - i))),
			UserId:      ownerId,
			ProjectId:   project.ID,
		}
		entries = append(entries, entry)
		err := handlerTest.TimeEntryUsecase.AddTimeEntry(&entry)
		assert.Nil(t, err)
	}
	return entries
}
