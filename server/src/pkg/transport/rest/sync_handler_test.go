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
