package usecase

import (
	"errors"
	"fmt"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_timeEntryUsecase_AddTimeEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, timeEntry.Description, entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, userId, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_AddTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   projectId,
	}
	err = usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var projectNotFoundError *ProjectNotFoundError
	assert.True(t, errors.As(err, &projectNotFoundError))
}

func Test_timeEntryUsecase_GetTimeEntryById(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entry, err := usecaseTest.TimeEntryUsecase.GetTimeEntryById(timeEntry.ID)
	assert.Nil(t, err)
	assert.Equal(t, timeEntry.Description, entry.Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entry.StartTime)
	assert.Equal(t, userId, entry.UserId)
	assert.Equal(t, project.ID, entry.ProjectId)
	assert.True(t, entry.EndTime.IsZero())
}

func Test_timeEntryUsecase_GetTimeEntryByIdFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = usecaseTest.TimeEntryUsecase.GetTimeEntryById(id)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_timeEntryUsecase_GetAllTimeEntriesOfUser(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	otherUserId := GetTestUserId(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	_ = addTimeEntries(t, usecaseTest.TimeEntryUsecase, 3, userId, project)
	_ = addTimeEntriesWithStartIndex(t, usecaseTest.TimeEntryUsecase, 4, 3, otherUserId, project)

	entriesOfUser, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	for i, entry := range entriesOfUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+1), entry.Description)
	}

	entriesOfOtherUser, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(otherUserId)
	assert.Nil(t, err)
	for i, entry := range entriesOfOtherUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+4), entry.Description)
	}
}

func Test_timeEntryUsecase_GetAllTimeEntriesOfUserAndProject(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)
	otherProject := addProject(t, usecaseTest.ProjectUsecase, "otherproject", userId)

	_ = addTimeEntries(t, usecaseTest.TimeEntryUsecase, 3, userId, project)
	_ = addTimeEntriesWithStartIndex(t, usecaseTest.TimeEntryUsecase, 4, 3, userId, otherProject)

	entriesOfUser, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUserAndProject(userId, project.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+1), entry.Description)
	}

	entriesOfOtherUser, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUserAndProject(userId, otherProject.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfOtherUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+4), entry.Description)
	}
}

func Test_timeEntryUsecase_UpdateTimeEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "updatedTimeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, userId, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		ID:          id,
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}

	timeEntry.Description = "updatedTimeentry"
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var notFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &notFoundError))

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entryList))
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfUserIdIsEmpty(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	timeEntry.UserId = uuid.Nil
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, userId, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfProjectIdIsEmpty(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	timeEntry.ProjectId = uuid.Nil
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, userId, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry.ProjectId = projectId
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var projectNotFoundError *ProjectNotFoundError
	assert.True(t, errors.As(err, &projectNotFoundError))

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, userId, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_DeleteTimeEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	err = usecaseTest.TimeEntryUsecase.DeleteTimeEntry(timeEntry.ID)
	assert.Nil(t, err)

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entryList))
	// Make sure that the associated entities are still there:
	_, err = usecaseTest.ProjectUsecase.GetProjectById(project.ID)
	assert.Nil(t, err)
}

func Test_timeEntryUsecase_DeleteTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = usecaseTest.TimeEntryUsecase.DeleteTimeEntry(notExistingId)
	assert.NotNil(t, err)
	var notFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &notFoundError))

	entryList, err := usecaseTest.TimeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
}

func assertTimesAreEqual(t *testing.T, time1 time.Time, time2 time.Time) {
	// We cannot check the milliseconds here because they get lost in the database:
	assert.Equal(t, time1.Hour(), time2.Hour())
	assert.Equal(t, time1.Minute(), time2.Minute())
	assert.Equal(t, time1.Second(), time2.Second())
}

func addTimeEntries(t *testing.T, timeEntryUsecase TimeEntryUsecase, count int, ownerId uuid.UUID, project model.Project) []model.TimeEntry {
	return addTimeEntriesWithStartIndex(t, timeEntryUsecase, 1, count, ownerId, project)
}

func addTimeEntriesWithStartIndex(t *testing.T, timeEntryUsecase TimeEntryUsecase, startIndex int, count int, ownerId uuid.UUID, project model.Project) []model.TimeEntry {
	var entries []model.TimeEntry
	startTime := time.Now()
	oneHour := 1000 * 1000 * 60 * 60 // duration is in nanoseconds
	for i := 0; i < count; i++ {
		entry := model.TimeEntry{
			Description: fmt.Sprintf("entry %v", startIndex+i),
			StartTime:   startTime.Add(time.Duration(oneHour * (count - i))),
			UserId:      ownerId,
			ProjectId:   project.ID,
		}
		entries = append(entries, entry)
		err := timeEntryUsecase.AddTimeEntry(&entry)
		assert.Nil(t, err)
	}
	return entries
}
