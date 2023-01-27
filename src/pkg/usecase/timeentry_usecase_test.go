package usecase

import (
	"testing"
	"time"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/stretchr/testify/assert"
)

func Test_timeEntryUsecase_AddTimeEntry(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, timeEntry.Description, entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func assertTimesAreEqual(t *testing.T, time1 time.Time, time2 time.Time) {
	// We cannot check the milliseconds here because they get lost in the database:
	assert.Equal(t, time1.Hour(), time2.Hour())
	assert.Equal(t, time1.Minute(), time2.Minute())
	assert.Equal(t, time1.Second(), time2.Second())
}
