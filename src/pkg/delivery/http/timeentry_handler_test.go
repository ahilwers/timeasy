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

	projectsFromDb, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}
