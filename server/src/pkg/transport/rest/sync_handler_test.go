package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_syncHandler_GetChangedTimeEntries(t *testing.T) {
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

	clientId := "test_client_id"

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	unchangedTimeEntry := model.TimeEntry{
		Description: "unchanged_timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&unchangedTimeEntry, userId, clientId)
	assert.Nil(t, err)

	updatedTimeEntry := model.TimeEntry{
		Description: "original_timeentry",
		StartTime:   startTime.Add(time.Hour),
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&updatedTimeEntry, userId, clientId)
	assert.Nil(t, err)
	updatedTimeEntry.Description = "updated_timeetry"
	err = handlerTest.TimeEntryUsecase.UpdateTimeEntry(&updatedTimeEntry, userId, clientId)
	assert.Nil(t, err)

	deletedTimeEntry := model.TimeEntry{
		Description: "deleted_timeentry",
		StartTime:   startTime.Add(time.Hour).Add(time.Hour),
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&deletedTimeEntry, userId, clientId)
	assert.Nil(t, err)
	err = handlerTest.TimeEntryUsecase.DeleteTimeEntry(deletedTimeEntry.ID, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	sinceChangeLogEntry := 3 // 1 is the project, 2 is the unchanged time entry and 3 is the added antry but we want the updated one.
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/sync/changed/%v", sinceChangeLogEntry), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var syncEntries SyncEntries
	err = json.Unmarshal(w.Body.Bytes(), &syncEntries)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(syncEntries.TimeEntries))

	foundDeleted := false
	foundUpdated := false

	for _, entry := range syncEntries.TimeEntries {
		if entry.ChangeType == DELETED {
			assert.Equal(t, deletedTimeEntry.Description, entry.Description)
			assert.Equal(t, deletedTimeEntry.ProjectId, syncEntries.TimeEntries[0].ProjectId)
			foundDeleted = true
		} else if entry.ChangeType == CHANGED {
			assert.Equal(t, updatedTimeEntry.Description, entry.Description)
			assert.Equal(t, updatedTimeEntry.ProjectId, syncEntries.TimeEntries[1].ProjectId)
			foundUpdated = true
		}
	}

	assert.True(t, foundDeleted, "Deleted time entry not found in response")
	assert.True(t, foundUpdated, "Updated time entry not found in response")

	// The new entry is correct because if we fetch from this entry on only the new entry is interesting
	// But for the deleted one only the deleted entry should be returned, not the new one
}

func Test_syncHandler_SendNewLocalTimeEntries(t *testing.T) {
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

	clientId := "test_client_id"

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)
	id, err := uuid.NewV4()
	assert.Nil(t, err)

	description := "timeEntry1"
	timeEntry1 := ChangedTimeEntryDto{
		Id:          id,
		Description: description,
		StartTime:   startTime.Format(time.RFC3339),
		EndTime:     endTime.Format(time.RFC3339),
		ProjectId:   project.ID,
		ChangeType:  NEW,
	}

	syncEntries := SyncEntries{
		TimeEntries: []ChangedTimeEntryDto{timeEntry1},
	}
	entryJson, err := json.Marshal(syncEntries)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	entryReader := bytes.NewReader(entryJson)
	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, id, entries[0].ID)
	assert.Equal(t, "timeEntry1", entries[0].Description)
	assert.Equal(t, startTime, entries[0].StartTime)
	assert.Equal(t, endTime, entries[0].EndTime)
	assert.Equal(t, project.ID, entries[0].ProjectId)
}

func Test_syncHandler_SendUpdatedLocalTimeEntries(t *testing.T) {
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

	clientId := "test_client_id"

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)

	// Create a time entry:
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		EndTime:     endTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry, userId, clientId)
	assert.Nil(t, err)

	// Now let's update the time entry:
	changeTime := time.Now().Add(time.Hour).UTC()
	description := "updatedTimeEntry"
	updatedTimeEntry := ChangedTimeEntryDto{
		Id:              timeEntry.ID,
		Description:     description,
		StartTime:       startTime.Format(time.RFC3339),
		EndTime:         endTime.Format(time.RFC3339),
		ProjectId:       project.ID,
		ChangeType:      CHANGED,
		ChangeTimestamp: changeTime.Format(time.RFC3339),
	}

	syncEntries := SyncEntries{
		TimeEntries: []ChangedTimeEntryDto{updatedTimeEntry},
	}
	entryJson, err := json.Marshal(syncEntries)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	entryReader := bytes.NewReader(entryJson)
	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, timeEntry.ID, entries[0].ID)
	assert.Equal(t, "updatedTimeEntry", entries[0].Description)
	assert.Equal(t, startTime, entries[0].StartTime)
	assert.Equal(t, endTime, entries[0].EndTime)
	assert.Equal(t, project.ID, entries[0].ProjectId)
}

func Test_syncHandler_SendUpdatedLocalTimeEntries_ShouldNotUpdateMissingFields(t *testing.T) {
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

	clientId := "test_client_id"

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)

	// Create a time entry:
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		EndTime:     endTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry, userId, clientId)
	assert.Nil(t, err)

	// Now let's update the time entry:
	changeTime := time.Now().Add(time.Hour).UTC()
	updatedTimeEntry := ChangedTimeEntryDto{
		Id:              timeEntry.ID,
		Description:     "",
		StartTime:       startTime.Format(time.RFC3339),
		EndTime:         endTime.Format(time.RFC3339),
		ProjectId:       project.ID,
		ChangeType:      CHANGED,
		ChangeTimestamp: changeTime.Format(time.RFC3339),
	}

	syncEntries := SyncEntries{
		TimeEntries: []ChangedTimeEntryDto{updatedTimeEntry},
	}
	entryJson, err := json.Marshal(syncEntries)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	entryReader := bytes.NewReader(entryJson)
	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entries))
	assert.Equal(t, timeEntry.ID, entries[0].ID)
	assert.Equal(t, "timeentry", entries[0].Description)
	assert.Equal(t, startTime, entries[0].StartTime)
	assert.Equal(t, endTime, entries[0].EndTime)
	assert.Equal(t, project.ID, entries[0].ProjectId)
}

func Test_syncHandler_SendDeletedLocalTimeEntries(t *testing.T) {
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

	clientId := "test_client_id"

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)

	// Create a time entry:
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		EndTime:     endTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry, userId, clientId)
	assert.Nil(t, err)

	// Now let's delete the time entry:
	changeTime := time.Now().Add(time.Hour).UTC()
	description := "deletedTimeEntry"
	deletedTimeEntry := ChangedTimeEntryDto{
		Id:              timeEntry.ID,
		Description:     description,
		StartTime:       startTime.Format(time.RFC3339),
		EndTime:         endTime.Format(time.RFC3339),
		ProjectId:       project.ID,
		ChangeType:      DELETED,
		ChangeTimestamp: changeTime.Format(time.RFC3339),
	}

	syncEntries := SyncEntries{
		TimeEntries: []ChangedTimeEntryDto{deletedTimeEntry},
	}
	entryJson, err := json.Marshal(syncEntries)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	entryReader := bytes.NewReader(entryJson)
	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entries))
}

func Test_syncHandler_GetChangedProjects(t *testing.T) {
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

	clientId := "test_client_id"

	unchangedProject := model.Project{
		Name:   "unchanged_project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&unchangedProject, userId, clientId)
	assert.Nil(t, err)

	updatedProject := model.Project{
		Name:   "original_project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&updatedProject, userId, clientId)
	assert.Nil(t, err)
	updatedProject.Name = "updated_timeetry"
	err = handlerTest.ProjectUsecase.UpdateProject(&updatedProject, userId, clientId)
	assert.Nil(t, err)

	deletedProject := model.Project{
		Name:   "deleted_project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&deletedProject, userId, clientId)
	assert.Nil(t, err)
	err = handlerTest.ProjectUsecase.DeleteProject(deletedProject.ID, userId, clientId)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/sync/changed/%v", 2), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var syncEntries SyncEntries
	err = json.Unmarshal(w.Body.Bytes(), &syncEntries)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(syncEntries.Projects))

	foundDeleted := false
	foundUpdated := false

	for _, poroject := range syncEntries.Projects {
		if poroject.ChangeType == DELETED {
			assert.Equal(t, deletedProject.Name, poroject.Name)
			foundDeleted = true
		} else if poroject.ChangeType == CHANGED {
			assert.Equal(t, updatedProject.Name, poroject.Name)
			foundUpdated = true
		}
	}

	assert.True(t, foundDeleted, "Deleted project not found in response")
	assert.True(t, foundUpdated, "Updated project not found in response")
}

func Test_syncHandler_SendUpdatedLocalProjects(t *testing.T) {
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
	clientId := "test_client_id"

	project := model.Project{
		Name:   "project",
		UserId: userId,
		Color:  "#ff0000",
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	// Now let's update the project
	changeTime := time.Now().Add(time.Hour).UTC()
	description := "updatedProject"
	deadline := model.NewDateOnly(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC))
	hourlyRate := 20.0
	timeBudget := 10
	color := "#00ff00"
	updatedProject := ChangedProjectDto{
		Id:              project.ID,
		Name:            description,
		Deadline:        &deadline,
		HourlyRate:      &hourlyRate,
		TimeBudget:      &timeBudget,
		Color:           &color,
		ChangeType:      CHANGED,
		ChangeTimestamp: changeTime.Format(time.RFC3339),
	}

	syncEntries := SyncEntries{
		Projects: []ChangedProjectDto{updatedProject},
	}
	entryJson, err := json.Marshal(syncEntries)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	entryReader := bytes.NewReader(entryJson)
	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projects, err := handlerTest.ProjectUsecase.GetAllProjectsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projects))
	assert.Equal(t, project.ID, projects[0].ID)
	assert.Equal(t, "updatedProject", projects[0].Name)
	assert.True(t, projects[0].HourlyRate.Equal(decimal.NewFromInt(20)))
	assert.True(t, deadline.Equal(projects[0].Deadline))
	assert.Equal(t, 10, projects[0].TimeBudget)
	assert.Equal(t, "#00ff00", projects[0].Color)
}

func Test_syncHandler_SendUpdatedLocalProjects_ShouldNotUpdateMissingFields(t *testing.T) {
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

	clientId := "test_client_id"

	deadline := model.NewDateOnly(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC))
	hourlyRate := decimal.NewFromInt(20)
	timeBudget := 10
	project := model.Project{
		Name:       "project",
		UserId:     userId,
		Color:      "#ff0000",
		Deadline:   deadline,
		HourlyRate: hourlyRate,
		TimeBudget: timeBudget,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project, userId, clientId)
	assert.Nil(t, err)

	// Now let's update the project
	changeTime := time.Now().Add(time.Hour).UTC()
	description := "updatedProject"
	updatedProject := ChangedProjectDto{
		Id:              project.ID,
		Name:            description,
		Color:           nil,
		Deadline:        nil,
		HourlyRate:      nil,
		TimeBudget:      nil,
		ChangeType:      CHANGED,
		ChangeTimestamp: changeTime.Format(time.RFC3339),
	}

	syncEntries := SyncEntries{
		Projects: []ChangedProjectDto{updatedProject},
	}
	entryJson, err := json.Marshal(syncEntries)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	entryReader := bytes.NewReader(entryJson)
	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projects, err := handlerTest.ProjectUsecase.GetAllProjectsOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projects))
	assert.Equal(t, project.ID, projects[0].ID)
	assert.Equal(t, "updatedProject", projects[0].Name)
	assert.True(t, projects[0].HourlyRate.Equal(decimal.NewFromInt(20)))
	assert.True(t, deadline.Equal(projects[0].Deadline))
	assert.Equal(t, 10, projects[0].TimeBudget)
	assert.Equal(t, "#ff0000", projects[0].Color)
}

func stringToTime(timeString string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeString)
	return t
}
