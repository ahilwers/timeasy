package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_timeEntryHandler_AddTimeEntry(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	project := model.Project{
		Name:   "project",
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"entry1", startTime.Unix(), project.ID))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	projectsFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "entry1", projectsFromDb[0].Description)
	assert.Equal(t, user.ID, projectsFromDb[0].UserId)
	assert.Equal(t, startTime, projectsFromDb[0].StartTime)
	assert.True(t, projectsFromDb[0].EndTime.IsZero())
}

func Test_timeEntryHandler_AddTimeEntryFailsIfProjectIdMissing(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})
	project := model.Project{
		Name:   "project",
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v}", "entry1", startTime.Unix()))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	projectsFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_timeEntryHandler_AddTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	w := httptest.NewRecorder()

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	missingProjectId, err := uuid.NewV4()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"entry1", startTime.Unix(), missingProjectId))
	req, err := http.NewRequest("POST", "/api/v1/timeentries", reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", missingProjectId))

	entriesFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entriesFromDb))
}

func Test_timeEntryHandler_UpdateTimeEntry(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)
	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	project := model.Project{
		Name:   "project",
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      user.ID,
	}
	err = TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), timeEntry.ProjectId))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))

	entriesFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "updatedentry", entriesFromDb[0].Description)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
	assert.Equal(t, timeEntry.ProjectId, entriesFromDb[0].ProjectId)
	assert.Equal(t, timeEntry.UserId, entriesFromDb[0].UserId)
}

func Test_timeEntryHandler_UpdateTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)
	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	project := model.Project{
		Name:   "project",
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), project.ID))
	missingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", missingId), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("entry with id %v not found", missingId))

	entriesFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entriesFromDb))
}

func Test_timeEntryHandler_UpdateTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)
	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	project := model.Project{
		Name:   "project",
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      user.ID,
	}
	err = TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	missingProjectId, err := uuid.NewV4()
	assert.Nil(t, err)
	reader := strings.NewReader(fmt.Sprintf("{\"description\": \"%v\", \"startTimeUTCUnix\": %v, \"projectId\": \"%v\"}",
		"updatedentry", startTime.Unix(), missingProjectId))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/timeentries/%v", timeEntry.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code, GetErrorMessageFromResponse(t, w.Body.Bytes()))
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", missingProjectId))

	entriesFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entriesFromDb))
	assert.Equal(t, "timeentry", entriesFromDb[0].Description)
	assert.Equal(t, startTime, entriesFromDb[0].StartTime)
	assert.True(t, entriesFromDb[0].EndTime.IsZero())
	assert.Equal(t, timeEntry.ProjectId, entriesFromDb[0].ProjectId)
	assert.Equal(t, timeEntry.UserId, entriesFromDb[0].UserId)
}
