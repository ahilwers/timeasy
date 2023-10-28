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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)

	unchangedTimeEntry := model.TimeEntry{
		Description: "unchanged_timeentry",
		StartTime:   startTime,
		ProjectId:   project.ID,
		UserId:      userId,
	}
	unchangedTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	unchangedTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&unchangedTimeEntry)
	assert.Nil(t, err)

	updatedTimeEntry := model.TimeEntry{
		Description: "original_timeentry",
		StartTime:   startTime.Add(time.Hour),
		ProjectId:   project.ID,
		UserId:      userId,
	}
	updatedTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	updatedTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&updatedTimeEntry)
	assert.Nil(t, err)
	updatedTimeEntry.Description = "updated_timeetry"
	err = handlerTest.TimeEntryUsecase.UpdateTimeEntry(&updatedTimeEntry)
	assert.Nil(t, err)

	deletedTimeEntry := model.TimeEntry{
		Description: "deleted_timeentry",
		StartTime:   startTime.Add(time.Hour).Add(time.Hour),
		ProjectId:   project.ID,
		UserId:      userId,
	}
	deletedTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	deletedTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&deletedTimeEntry)
	assert.Nil(t, err)
	err = handlerTest.TimeEntryUsecase.DeleteTimeEntry(deletedTimeEntry.ID)
	assert.Nil(t, err)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/sync/changed/%v", time.Now().Unix()), nil)
	handlerTest.Router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var syncEntries []ChangedTimeEntryDto
	json.Unmarshal(w.Body.Bytes(), &syncEntries)
	assert.Equal(t, 2, len(syncEntries))

	assert.Equal(t, deletedTimeEntry.Description, syncEntries[0].Description)
	assert.Equal(t, deletedTimeEntry.StartTime, time.Unix(syncEntries[0].StartTimeUTCUnix, 0).UTC())
	assert.Equal(t, deletedTimeEntry.EndTime, time.Unix(syncEntries[0].EndTimeUTCUnix, 0).UTC())
	assert.Equal(t, deletedTimeEntry.ProjectId, syncEntries[0].ProjectId)
	assert.Equal(t, DELETED, syncEntries[0].ChangeType)

	assert.Equal(t, updatedTimeEntry.Description, syncEntries[1].Description)
	assert.Equal(t, updatedTimeEntry.StartTime, time.Unix(syncEntries[1].StartTimeUTCUnix, 0).UTC())
	assert.Equal(t, updatedTimeEntry.EndTime, time.Unix(syncEntries[1].EndTimeUTCUnix, 0).UTC())
	assert.Equal(t, updatedTimeEntry.ProjectId, syncEntries[1].ProjectId)
	assert.Equal(t, CHANGED, syncEntries[1].ChangeType)

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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	startTime := time.Date(2023, 1, 28, 11, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 1, 28, 11, 1, 0, 0, time.UTC)
	id, err := uuid.NewV4()
	assert.Nil(t, err)

	timeEntry1 := ChangedTimeEntryDto{
		Id:               id,
		Description:      "timeEntry1",
		StartTimeUTCUnix: startTime.Unix(),
		EndTimeUTCUnix:   endTime.Unix(),
		ProjectId:        project.ID,
		ChangeType:       NEW,
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
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
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	// Now let's update the time entry:
	changeTime := time.Now().Add(time.Hour).UTC()
	updatedTimeEntry := ChangedTimeEntryDto{
		Id:                     timeEntry.ID,
		Description:            "updatedTimeEntry",
		StartTimeUTCUnix:       startTime.Unix(),
		EndTimeUTCUnix:         endTime.Unix(),
		ProjectId:              project.ID,
		ChangeType:             CHANGED,
		ChangeTimestampUTCUnix: changeTime.Unix(),
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

	project := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err = handlerTest.ProjectUsecase.AddProject(&project)
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
	err = handlerTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	// Now let's delete the time entry:
	changeTime := time.Now().Add(time.Hour).UTC()
	deletedTimeEntry := ChangedTimeEntryDto{
		Id:                     timeEntry.ID,
		Description:            "deletedTimeEntry",
		StartTimeUTCUnix:       startTime.Unix(),
		EndTimeUTCUnix:         endTime.Unix(),
		ProjectId:              project.ID,
		ChangeType:             DELETED,
		ChangeTimestampUTCUnix: changeTime.Unix(),
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
