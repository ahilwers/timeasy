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

	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"entry1", startTime.Format(time.RFC3339), project.ID))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	assert.Nil(t, err)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	entriesFromDb, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "entry1", entriesFromDb[0].Description)
	assert.Equal(t, userId, entriesFromDb[0].UserId)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
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

	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\"}", "entry1", startTime.Format(time.RFC3339)))
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
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"entry1", startTime.Format(time.RFC3339), missingProjectId))
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
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"updatedentry", startTime.Format(time.RFC3339), timeEntry.ProjectId))
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
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"updatedentry", startTime.Format(time.RFC3339), project.ID))
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
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"updatedentry", startTime.Format(time.RFC3339), missingProjectId))
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
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"updatedentry", startTime.Format(time.RFC3339), timeEntry.ProjectId))
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
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTime\": \"%v\", \"projectId\": \"%v\"}",
		"updatedentry", startTime.Format(time.RFC3339), timeEntry.ProjectId))
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
	Id          uuid.UUID
	Description string    `json:"description" binding:"required"`
	StartTime   string    `json:"startTime" binding:"required"`
	EndTime     string    `json:"endTime,omitempty"`
	ProjectId   uuid.UUID `json:"projectId" binding:"required"`
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
	startTimeFromEntry, err := time.Parse(time.RFC3339, entryFromService.StartTime)
	assert.Nil(t, err)
	assert.Equal(t, startTime, startTimeFromEntry)
	assert.Equal(t, "", entryFromService.EndTime)
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
	startTimeFromEntry, err := time.Parse(time.RFC3339, entryFromService.StartTime)
	assert.Nil(t, err)
	assert.Equal(t, startTime, startTimeFromEntry)
	assert.Equal(t, "", entryFromService.EndTime)
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

func Test_timeEntryHandler_GetAllTimeEntries_WithProjectId_ReturnsTimeEntriesOfProject(t *testing.T) {
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

	project1 := model.Project{
		Name:   "project1",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project1)
	assert.Nil(t, err)
	project2 := model.Project{
		Name:   "project2",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project2)
	assert.Nil(t, err)

	addTimeEntries(t, handlerTest, 3, userId, project1)
	addTimeEntries(t, handlerTest, 2, userId, project2)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/timeentries?projectId="+project1.ID.String(), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entriesFromService []timeEntryDto
	json.Unmarshal(w.Body.Bytes(), &entriesFromService)
	assert.Equal(t, 3, len(entriesFromService))
	for index, entryFromService := range entriesFromService {
		assert.Equal(t, fmt.Sprintf("entry %v", index+1), entryFromService.Description)
		assert.Equal(t, project1.ID, entryFromService.ProjectId)
	}
}

func Test_timeEntryHandler_GetAllTimeEntries_WithDateRangeAndProjectId_ReturnsTimeEntriesOfProjectWithinThisRange(t *testing.T) {
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

	project1 := model.Project{
		Name:   "project1",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project1)
	assert.Nil(t, err)
	project2 := model.Project{
		Name:   "project2",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project2)
	assert.Nil(t, err)

	var startTime1 = time.Date(2025, 3, 10, 11, 0, 0, 0, time.UTC)
	addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, 1, 3, userId, project1, startTime1)
	addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, 1, 3, userId, project2, startTime1)
	var startTime2 = time.Date(2025, 3, 16, 17, 0, 0, 0, time.UTC)
	addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, 4, 2, userId, project1, startTime2)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/timeentries?projectId="+project1.ID.String()+"&startDate=2025-03-10&endDate=2025-03-15", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entriesFromService []timeEntryDto
	json.Unmarshal(w.Body.Bytes(), &entriesFromService)
	assert.Equal(t, 3, len(entriesFromService))
	for index, entryFromService := range entriesFromService {
		assert.Equal(t, fmt.Sprintf("entry %v", index+1), entryFromService.Description)
		assert.Equal(t, project1.ID, entryFromService.ProjectId)
		timeEntryStartTime, convertError := time.Parse(time.RFC3339, entryFromService.StartTime)
		assert.Nil(t, convertError)
		assert.True(t, timeEntryStartTime.After(startTime1) && timeEntryStartTime.Before(startTime2))
	}
}

func Test_timeEntryHandler_GetAllTimeEntries_WithDateRange_ReturnsTimeEntriesOfAllProjectsWithinThisRange(t *testing.T) {
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

	project1 := model.Project{
		Name:   "project1",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project1)
	assert.Nil(t, err)
	project2 := model.Project{
		Name:   "project2",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project2)
	assert.Nil(t, err)

	var startTime1 = time.Date(2025, 3, 10, 11, 0, 0, 0, time.UTC)
	addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, 1, 3, userId, project1, startTime1)
	addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, 1, 3, userId, project2, startTime1)
	var startTime2 = time.Date(2025, 3, 16, 17, 0, 0, 0, time.UTC)
	addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, 4, 2, userId, project1, startTime2)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/timeentries?&startDate=2025-03-10&endDate=2025-03-15", nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var entriesFromService []timeEntryDto
	json.Unmarshal(w.Body.Bytes(), &entriesFromService)
	assert.Equal(t, 6, len(entriesFromService))
}

func addTimeEntries(t *testing.T, handlerTest *HandlerTest, count int, ownerId uuid.UUID, project model.Project) []model.TimeEntry {
	return addTimeEntriesWithStartIndex(t, handlerTest, 1, count, ownerId, project)
}

func addTimeEntriesWithStartIndex(t *testing.T, handlerTest *HandlerTest, startIndex int, count int, ownerId uuid.UUID, project model.Project) []model.TimeEntry {
	return addTimeEntriesWithStartIndexAndStartTime(t, handlerTest, startIndex, count, ownerId, project, time.Now())
}

func addTimeEntriesWithStartIndexAndStartTime(t *testing.T, handlerTest *HandlerTest, startIndex int, count int, ownerId uuid.UUID, project model.Project, startTime time.Time) []model.TimeEntry {
	var entries []model.TimeEntry
	oneHour := 1000 * 1000 * 60 * 60 // duration is in nanoseconds
	oneHourAndThirtyMinutes := oneHour + 1000*1000*30*60
	for i := 0; i < count; i++ {
		entry := model.TimeEntry{
			Description: fmt.Sprintf("entry %v", startIndex+i),
			StartTime:   startTime.Add(time.Duration(oneHour * (count - i))),
			EndTime:     startTime.Add(time.Duration(oneHourAndThirtyMinutes * (count - i))),
			UserId:      ownerId,
			ProjectId:   project.ID,
		}
		entries = append(entries, entry)
		err := handlerTest.TimeEntryUsecase.AddTimeEntry(&entry)
		assert.Nil(t, err)
	}
	return entries
}
