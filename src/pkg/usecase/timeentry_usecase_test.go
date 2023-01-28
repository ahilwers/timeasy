package usecase

import (
	"errors"
	"fmt"
	"testing"
	"time"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/gofrs/uuid"
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

func Test_timeEntryUsecase_GetTimeEntryById(t *testing.T) {
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

	entry, err := timeEntryUsecase.GetTimeEntryById(timeEntry.ID)
	assert.Nil(t, err)
	assert.Equal(t, timeEntry.Description, entry.Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entry.StartTime)
	assert.Equal(t, user.ID, entry.UserId)
	assert.Equal(t, project.ID, entry.ProjectId)
	assert.True(t, entry.EndTime.IsZero())
}

func Test_timeEntryUsecase_GetTimeEntryByIdFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = timeEntryUsecase.GetTimeEntryById(id)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_timeEntryUsecase_GetAllTimeEntriesOfUser(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	otherUser := addUser(t, "otheruser", "otherpassword", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo)

	_ = addTimeEntries(t, timeEntryUsecase, 3, user, project)
	_ = addTimeEntriesWithStartIndex(t, timeEntryUsecase, 4, 3, otherUser, project)

	entriesOfUser, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+1), entry.Description)
	}

	entriesOfOtherUser, err := timeEntryUsecase.GetAllTimeEntriesOfUser(otherUser.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfOtherUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+4), entry.Description)
	}
}

func Test_timeEntryUsecase_GetAllTimeEntriesOfUserAndProject(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)
	otherProject := addProject(t, projectUsecase, "otherproject", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo)

	_ = addTimeEntries(t, timeEntryUsecase, 3, user, project)
	_ = addTimeEntriesWithStartIndex(t, timeEntryUsecase, 4, 3, user, otherProject)

	entriesOfUser, err := timeEntryUsecase.GetAllTimeEntriesOfUserAndProject(user.ID, project.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+1), entry.Description)
	}

	entriesOfOtherUser, err := timeEntryUsecase.GetAllTimeEntriesOfUserAndProject(user.ID, otherProject.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfOtherUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+4), entry.Description)
	}
}

func assertTimesAreEqual(t *testing.T, time1 time.Time, time2 time.Time) {
	// We cannot check the milliseconds here because they get lost in the database:
	assert.Equal(t, time1.Hour(), time2.Hour())
	assert.Equal(t, time1.Minute(), time2.Minute())
	assert.Equal(t, time1.Second(), time2.Second())
}

func addTimeEntries(t *testing.T, usecase TimeEntryUsecase, count int, owner model.User, project model.Project) []model.TimeEntry {
	return addTimeEntriesWithStartIndex(t, usecase, 1, count, owner, project)
}

func addTimeEntriesWithStartIndex(t *testing.T, usecase TimeEntryUsecase, startIndex int, count int, owner model.User, project model.Project) []model.TimeEntry {
	var entries []model.TimeEntry
	for i := 0; i < count; i++ {
		entry := model.TimeEntry{
			Description: fmt.Sprintf("entry %v", startIndex+i),
			StartTime:   time.Now(),
			UserId:      owner.ID,
			ProjectId:   project.ID,
		}
		entries = append(entries, entry)
		err := usecase.AddTimeEntry(&entry)
		assert.Nil(t, err)
	}
	return entries
}
