package rest

//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"github.com/shopspring/decimal"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//	"time"
//	"timeasy-server/pkg/domain/model"
//
//	"github.com/gofrs/uuid"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//)
//
//func Test_syncHandler_GetChangedTimeEntries(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	project := model.Project{
//		Name:   "project",
//		UserId: userId,
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
//
//	unchangedTimeEntry := model.TimeEntry{
//		Description: "unchanged_timeentry",
//		StartTime:   startTime,
//		ProjectId:   project.ID,
//		UserId:      userId,
//	}
//	unchangedTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	unchangedTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&unchangedTimeEntry)
//	assert.Nil(t, err)
//
//	updatedTimeEntry := model.TimeEntry{
//		Description: "original_timeentry",
//		StartTime:   startTime.Add(time.Hour),
//		ProjectId:   project.ID,
//		UserId:      userId,
//	}
//	updatedTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	updatedTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&updatedTimeEntry)
//	assert.Nil(t, err)
//	updatedTimeEntry.Description = "updated_timeetry"
//	err = handlerTest.TimeEntryUsecase.UpdateTimeEntry(&updatedTimeEntry)
//	assert.Nil(t, err)
//
//	deletedTimeEntry := model.TimeEntry{
//		Description: "deleted_timeentry",
//		StartTime:   startTime.Add(time.Hour).Add(time.Hour),
//		ProjectId:   project.ID,
//		UserId:      userId,
//	}
//	deletedTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	deletedTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&deletedTimeEntry)
//	assert.Nil(t, err)
//	err = handlerTest.TimeEntryUsecase.DeleteTimeEntry(deletedTimeEntry.ID)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/sync/changed/%v", time.Now().UTC().Unix()), nil)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	var syncEntries SyncEntries
//	err = json.Unmarshal(w.Body.Bytes(), &syncEntries)
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(syncEntries.TimeEntries))
//
//	assert.Equal(t, deletedTimeEntry.Description, *syncEntries.TimeEntries[0].Description)
//	assert.Equal(t, deletedTimeEntry.StartTime, stringToTime(syncEntries.TimeEntries[0].StartTime))
//	assert.Equal(t, deletedTimeEntry.EndTime, stringToTime(syncEntries.TimeEntries[0].EndTime))
//	assert.Equal(t, deletedTimeEntry.ProjectId, syncEntries.TimeEntries[0].ProjectId)
//	assert.Equal(t, DELETED, syncEntries.TimeEntries[0].Operation)
//
//	assert.Equal(t, updatedTimeEntry.Description, *syncEntries.TimeEntries[1].Description)
//	assert.Equal(t, updatedTimeEntry.StartTime, stringToTime(syncEntries.TimeEntries[1].StartTime))
//	assert.Equal(t, updatedTimeEntry.EndTime, stringToTime(syncEntries.TimeEntries[1].EndTime))
//	assert.Equal(t, updatedTimeEntry.ProjectId, syncEntries.TimeEntries[1].ProjectId)
//	assert.Equal(t, CHANGED, syncEntries.TimeEntries[1].Operation)
//}
//
//func Test_syncHandler_SendNewLocalTimeEntries(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	project := model.Project{
//		Name:   "project",
//		UserId: userId,
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
//	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)
//	id, err := uuid.NewV4()
//	assert.Nil(t, err)
//
//	description := "timeEntry1"
//	timeEntry1 := ChangedTimeEntryDto{
//		Id:          id,
//		Description: &description,
//		StartTime:   startTime.Format(time.RFC3339),
//		EndTime:     endTime.Format(time.RFC3339),
//		ProjectId:   project.ID,
//		Operation:  NEW,
//	}
//
//	syncEntries := SyncEntries{
//		TimeEntries: []ChangedTimeEntryDto{timeEntry1},
//	}
//	entryJson, err := json.Marshal(syncEntries)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	entryReader := bytes.NewReader(entryJson)
//	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(entries))
//	assert.Equal(t, id, entries[0].ID)
//	assert.Equal(t, "timeEntry1", entries[0].Description)
//	assert.Equal(t, startTime, entries[0].StartTime)
//	assert.Equal(t, endTime, entries[0].EndTime)
//	assert.Equal(t, project.ID, entries[0].ProjectId)
//}
//
//func Test_syncHandler_SendUpdatedLocalTimeEntries(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	project := model.Project{
//		Name:   "project",
//		UserId: userId,
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
//	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)
//
//	// Create a time entry:
//	timeEntry := model.TimeEntry{
//		Description: "timeentry",
//		StartTime:   startTime,
//		EndTime:     endTime,
//		ProjectId:   project.ID,
//		UserId:      userId,
//	}
//	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
//	assert.Nil(t, err)
//
//	// Now let's update the time entry:
//	changeTime := time.Now().Add(time.Hour).UTC()
//	description := "updatedTimeEntry"
//	updatedTimeEntry := ChangedTimeEntryDto{
//		Id:              timeEntry.ID,
//		Description:     &description,
//		StartTime:       startTime.Format(time.RFC3339),
//		EndTime:         endTime.Format(time.RFC3339),
//		ProjectId:       project.ID,
//		Operation:      CHANGED,
//		ChangeTimestamp: changeTime.Format(time.RFC3339),
//	}
//
//	syncEntries := SyncEntries{
//		TimeEntries: []ChangedTimeEntryDto{updatedTimeEntry},
//	}
//	entryJson, err := json.Marshal(syncEntries)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	entryReader := bytes.NewReader(entryJson)
//	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(entries))
//	assert.Equal(t, timeEntry.ID, entries[0].ID)
//	assert.Equal(t, "updatedTimeEntry", entries[0].Description)
//	assert.Equal(t, startTime, entries[0].StartTime)
//	assert.Equal(t, endTime, entries[0].EndTime)
//	assert.Equal(t, project.ID, entries[0].ProjectId)
//}
//
//func Test_syncHandler_SendUpdatedLocalTimeEntries_ShouldNotUpdateMissingFields(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	project := model.Project{
//		Name:   "project",
//		UserId: userId,
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
//	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)
//
//	// Create a time entry:
//	timeEntry := model.TimeEntry{
//		Description: "timeentry",
//		StartTime:   startTime,
//		EndTime:     endTime,
//		ProjectId:   project.ID,
//		UserId:      userId,
//	}
//	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
//	assert.Nil(t, err)
//
//	// Now let's update the time entry:
//	changeTime := time.Now().Add(time.Hour).UTC()
//	updatedTimeEntry := ChangedTimeEntryDto{
//		Id:              timeEntry.ID,
//		Description:     nil,
//		StartTime:       startTime.Format(time.RFC3339),
//		EndTime:         endTime.Format(time.RFC3339),
//		ProjectId:       project.ID,
//		Operation:      CHANGED,
//		ChangeTimestamp: changeTime.Format(time.RFC3339),
//	}
//
//	syncEntries := SyncEntries{
//		TimeEntries: []ChangedTimeEntryDto{updatedTimeEntry},
//	}
//	entryJson, err := json.Marshal(syncEntries)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	entryReader := bytes.NewReader(entryJson)
//	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(entries))
//	assert.Equal(t, timeEntry.ID, entries[0].ID)
//	assert.Equal(t, "timeentry", entries[0].Description)
//	assert.Equal(t, startTime, entries[0].StartTime)
//	assert.Equal(t, endTime, entries[0].EndTime)
//	assert.Equal(t, project.ID, entries[0].ProjectId)
//}
//
//func Test_syncHandler_SendDeletedLocalTimeEntries(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	project := model.Project{
//		Name:   "project",
//		UserId: userId,
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
//	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)
//
//	// Create a time entry:
//	timeEntry := model.TimeEntry{
//		Description: "timeentry",
//		StartTime:   startTime,
//		EndTime:     endTime,
//		ProjectId:   project.ID,
//		UserId:      userId,
//	}
//	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
//	assert.Nil(t, err)
//
//	// Now let's delete the time entry:
//	changeTime := time.Now().Add(time.Hour).UTC()
//	description := "deletedTimeEntry"
//	deletedTimeEntry := ChangedTimeEntryDto{
//		Id:              timeEntry.ID,
//		Description:     &description,
//		StartTime:       startTime.Format(time.RFC3339),
//		EndTime:         endTime.Format(time.RFC3339),
//		ProjectId:       project.ID,
//		Operation:      DELETED,
//		ChangeTimestamp: changeTime.Format(time.RFC3339),
//	}
//
//	syncEntries := SyncEntries{
//		TimeEntries: []ChangedTimeEntryDto{deletedTimeEntry},
//	}
//	entryJson, err := json.Marshal(syncEntries)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	entryReader := bytes.NewReader(entryJson)
//	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	entries, err := handlerTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
//	assert.Nil(t, err)
//	assert.Equal(t, 0, len(entries))
//}
//
//func Test_syncHandler_GetChangedProjects(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	unchangedProject := model.Project{
//		Name:   "unchanged_project",
//		UserId: userId,
//	}
//	unchangedProject.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	unchangedProject.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	err = handlerTest.ProjectUsecase.AddProject(&unchangedProject)
//	assert.Nil(t, err)
//
//	updatedProject := model.Project{
//		Name:   "original_project",
//		UserId: userId,
//	}
//	updatedProject.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	updatedProject.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	err = handlerTest.ProjectUsecase.AddProject(&updatedProject)
//	assert.Nil(t, err)
//	updatedProject.Name = "updated_timeetry"
//	err = handlerTest.ProjectUsecase.UpdateProject(&updatedProject)
//	assert.Nil(t, err)
//
//	deletedProject := model.Project{
//		Name:   "deleted_project",
//		UserId: userId,
//	}
//	deletedProject.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	deletedProject.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
//	err = handlerTest.ProjectUsecase.AddProject(&deletedProject)
//	assert.Nil(t, err)
//	err = handlerTest.ProjectUsecase.DeleteProject(deletedProject.ID)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/sync/changed/%v", time.Now().UTC().Unix()), nil)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	var syncEntries SyncEntries
//	err = json.Unmarshal(w.Body.Bytes(), &syncEntries)
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(syncEntries.Projects))
//
//	assert.Equal(t, deletedProject.Name, syncEntries.Projects[0].Name)
//	assert.Equal(t, DELETED, syncEntries.Projects[0].Operation)
//
//	assert.Equal(t, updatedProject.Name, syncEntries.Projects[1].Name)
//	assert.Equal(t, CHANGED, syncEntries.Projects[1].Operation)
//}
//
//func Test_syncHandler_SendUpdatedLocalProjects(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	project := model.Project{
//		Name:   "project",
//		UserId: userId,
//		Color:  "#ff0000",
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	// Now let's update the project
//	changeTime := time.Now().Add(time.Hour).UTC()
//	description := "updatedProject"
//	deadline := model.NewDateOnly(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC))
//	hourlyRate := decimal.NewFromInt(20)
//	timeBudget := 10
//	color := "#00ff00"
//	updatedProject := ChangedProjectDto{
//		Id:              project.ID,
//		Name:            description,
//		Deadline:        &deadline,
//		HourlyRate:      &hourlyRate,
//		TimeBudget:      &timeBudget,
//		Color:           &color,
//		Operation:      CHANGED,
//		ChangeTimestamp: changeTime.Format(time.RFC3339),
//	}
//
//	syncEntries := SyncEntries{
//		Projects: []ChangedProjectDto{updatedProject},
//	}
//	entryJson, err := json.Marshal(syncEntries)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	entryReader := bytes.NewReader(entryJson)
//	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	projects, err := handlerTest.ProjectUsecase.GetAllProjectsOfUser(userId)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(projects))
//	assert.Equal(t, project.ID, projects[0].ID)
//	assert.Equal(t, "updatedProject", projects[0].Name)
//	assert.True(t, projects[0].HourlyRate.Equal(decimal.NewFromInt(20)))
//	assert.Equal(t, deadline.ToTime(), projects[0].Deadline.ToTime())
//	assert.Equal(t, 10, projects[0].TimeBudget)
//	assert.Equal(t, "#00ff00", projects[0].Color)
//}
//
//func Test_syncHandler_SendUpdatedLocalProjects_ShouldNotUpdateMissingFields(t *testing.T) {
//	userId, err := uuid.NewV4()
//	assert.Nil(t, err)
//	token := authTokenMock{}
//	token.On("GetUserId").Return(userId, nil)
//	token.On("HasRole", model.RoleUser).Return(true, nil)
//	token.On("HasRole", model.RoleAdmin).Return(false, nil)
//
//	verifier := tokenVerifierMock{}
//	verifier.On("VerifyToken", mock.Anything).Return(&token, nil)
//
//	handlerTest := NewHandlerTest(&verifier)
//	teardownTest := handlerTest.SetupTest(t)
//	defer teardownTest(t)
//
//	deadline := model.NewDateOnly(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC))
//	hourlyRate := decimal.NewFromInt(20)
//	timeBudget := 10
//	project := model.Project{
//		Name:       "project",
//		UserId:     userId,
//		Color:      "#ff0000",
//		Deadline:   deadline,
//		HourlyRate: hourlyRate,
//		TimeBudget: timeBudget,
//	}
//	err = handlerTest.ProjectUsecase.AddProject(&project)
//	assert.Nil(t, err)
//
//	// Now let's update the project
//	changeTime := time.Now().Add(time.Hour).UTC()
//	description := "updatedProject"
//	updatedProject := ChangedProjectDto{
//		Id:              project.ID,
//		Name:            description,
//		Color:           nil,
//		Deadline:        nil,
//		HourlyRate:      nil,
//		TimeBudget:      nil,
//		Operation:      CHANGED,
//		ChangeTimestamp: changeTime.Format(time.RFC3339),
//	}
//
//	syncEntries := SyncEntries{
//		Projects: []ChangedProjectDto{updatedProject},
//	}
//	entryJson, err := json.Marshal(syncEntries)
//	assert.Nil(t, err)
//
//	w := httptest.NewRecorder()
//
//	entryReader := bytes.NewReader(entryJson)
//	req, _ := http.NewRequest("POST", "/api/v1/sync/changed", entryReader)
//	handlerTest.Router.ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
//
//	projects, err := handlerTest.ProjectUsecase.GetAllProjectsOfUser(userId)
//	assert.Nil(t, err)
//	assert.Equal(t, 1, len(projects))
//	assert.Equal(t, project.ID, projects[0].ID)
//	assert.Equal(t, "updatedProject", projects[0].Name)
//	assert.True(t, projects[0].HourlyRate.Equal(decimal.NewFromInt(20)))
//	assert.Equal(t, deadline.ToTime(), projects[0].Deadline.ToTime())
//	assert.Equal(t, 10, projects[0].TimeBudget)
//	assert.Equal(t, "#ff0000", projects[0].Color)
//}
//
//func stringToTime(timeString string) time.Time {
//	t, _ := time.Parse(time.RFC3339, timeString)
//	return t
//}
