package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type weeklyStatisticsTestDto struct {
	Days []dailyStatisticsDto `json:"days"`
}

type dailyStatisticsTestDto struct {
	Weekday       string `json:"weekday"`
	TimeInSeconds int    `json:"timeInSeconds"`
}

func Test_WeeklyStatisticsHandler_GetWeeklyStatisticsReturnsAllDataFromProject(t *testing.T) {
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
	clientId := "1"
	err = handlerTest.ProjectUsecase.AddProject(&project1, userId, clientId)
	assert.Nil(t, err)
	project2 := model.Project{
		Name:   "project2",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project2, userId, clientId)
	assert.Nil(t, err)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 15, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Thursday
	startTime = time.Date(2025, time.January, 2, 11, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 2, 11, 45, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Sunday
	startTime = time.Date(2025, time.January, 5, 12, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 5, 13, 0, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Monday in another project1
	startTime = time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 30, 15, 15, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project2.ID, startTime, endTime)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/weeklystatistics/%v/%v?project=%v", 1, 2025, project1.ID), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var statistics weeklyStatisticsTestDto
	json.Unmarshal(w.Body.Bytes(), &statistics)
	assert.Equal(t, 7, len(statistics.Days))
	assert.Equal(t, 900, statistics.Days[0].TimeInSeconds)
	assert.Equal(t, "Monday", statistics.Days[0].Weekday)
	assert.Equal(t, 1800, statistics.Days[1].TimeInSeconds)
	assert.Equal(t, "Tuesday", statistics.Days[1].Weekday)
	assert.Equal(t, 0, statistics.Days[2].TimeInSeconds)
	assert.Equal(t, "Wednesday", statistics.Days[2].Weekday)
	assert.Equal(t, 2700, statistics.Days[3].TimeInSeconds)
	assert.Equal(t, "Thursday", statistics.Days[3].Weekday)
	assert.Equal(t, 0, statistics.Days[4].TimeInSeconds)
	assert.Equal(t, "Friday", statistics.Days[4].Weekday)
	assert.Equal(t, 0, statistics.Days[5].TimeInSeconds)
	assert.Equal(t, "Saturday", statistics.Days[5].Weekday)
	assert.Equal(t, 3600, statistics.Days[6].TimeInSeconds)
	assert.Equal(t, "Sunday", statistics.Days[6].Weekday)
}

func Test_WeeklyStatisticsHandler_GetWeeklyStatisticsWithoutProjectIdReturnsAllDataFromAllProjects(t *testing.T) {
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
	clientId := "1"
	err = handlerTest.ProjectUsecase.AddProject(&project1, userId, clientId)
	assert.Nil(t, err)
	project2 := model.Project{
		Name:   "project2",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project2, userId, clientId)
	assert.Nil(t, err)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 15, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Thursday
	startTime = time.Date(2025, time.January, 2, 11, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 2, 11, 45, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Sunday
	startTime = time.Date(2025, time.January, 5, 12, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 5, 13, 0, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project1.ID, startTime, endTime)
	// Monday in another project1
	startTime = time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 30, 15, 15, 0, 0, time.UTC)
	AddTimeEntry(t, handlerTest, userId, project2.ID, startTime, endTime)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/weeklystatistics/%v/%v", 1, 2025), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var statistics weeklyStatisticsTestDto
	json.Unmarshal(w.Body.Bytes(), &statistics)
	assert.Equal(t, 7, len(statistics.Days))
	assert.Equal(t, 1800, statistics.Days[0].TimeInSeconds)
	assert.Equal(t, "Monday", statistics.Days[0].Weekday)
	assert.Equal(t, 1800, statistics.Days[1].TimeInSeconds)
	assert.Equal(t, "Tuesday", statistics.Days[1].Weekday)
	assert.Equal(t, 0, statistics.Days[2].TimeInSeconds)
	assert.Equal(t, "Wednesday", statistics.Days[2].Weekday)
	assert.Equal(t, 2700, statistics.Days[3].TimeInSeconds)
	assert.Equal(t, "Thursday", statistics.Days[3].Weekday)
	assert.Equal(t, 0, statistics.Days[4].TimeInSeconds)
	assert.Equal(t, "Friday", statistics.Days[4].Weekday)
	assert.Equal(t, 0, statistics.Days[5].TimeInSeconds)
	assert.Equal(t, "Saturday", statistics.Days[5].Weekday)
	assert.Equal(t, 3600, statistics.Days[6].TimeInSeconds)
	assert.Equal(t, "Sunday", statistics.Days[6].Weekday)
}

func AddTimeEntry(t *testing.T, handlerTest *HandlerTest, userId uuid.UUID, projectId uuid.UUID, startTime time.Time, endTime time.Time) *model.TimeEntry {
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		EndTime:     endTime,
		UserId:      userId,
		ProjectId:   projectId,
	}
	clientId := "1"
	err := handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry, userId, clientId)
	assert.Nil(t, err)
	return &timeEntry
}
