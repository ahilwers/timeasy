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
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, timeEntry.Description, entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_AddTimeEntryFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	userId, err := uuid.NewV4()
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err = TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var userNotFoundError *UserNotFoundError
	assert.True(t, errors.As(err, &userNotFoundError))
}

func Test_timeEntryUsecase_AddTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   projectId,
	}
	err = TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var projectNotFoundError *ProjectNotFoundError
	assert.True(t, errors.As(err, &projectNotFoundError))
}

func Test_timeEntryUsecase_GetTimeEntryById(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entry, err := TestTimeEntryUsecase.GetTimeEntryById(timeEntry.ID)
	assert.Nil(t, err)
	assert.Equal(t, timeEntry.Description, entry.Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entry.StartTime)
	assert.Equal(t, user.ID, entry.UserId)
	assert.Equal(t, project.ID, entry.ProjectId)
	assert.True(t, entry.EndTime.IsZero())
}

func Test_timeEntryUsecase_GetTimeEntryByIdFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = TestTimeEntryUsecase.GetTimeEntryById(id)
	assert.NotNil(t, err)
	var entityNotFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &entityNotFoundError))
}

func Test_timeEntryUsecase_GetAllTimeEntriesOfUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	otherUser := addUser(t, "otheruser", "otherpassword", model.RoleList{model.RoleUser})

	project := addProject(t, "project", user)

	_ = addTimeEntries(t, 3, user, project)
	_ = addTimeEntriesWithStartIndex(t, 4, 3, otherUser, project)

	entriesOfUser, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+1), entry.Description)
	}

	entriesOfOtherUser, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(otherUser.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfOtherUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+4), entry.Description)
	}
}

func Test_timeEntryUsecase_GetAllTimeEntriesOfUserAndProject(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})

	project := addProject(t, "project", user)
	otherProject := addProject(t, "otherproject", user)

	_ = addTimeEntries(t, 3, user, project)
	_ = addTimeEntriesWithStartIndex(t, 4, 3, user, otherProject)

	entriesOfUser, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUserAndProject(user.ID, project.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+1), entry.Description)
	}

	entriesOfOtherUser, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUserAndProject(user.ID, otherProject.ID)
	assert.Nil(t, err)
	for i, entry := range entriesOfOtherUser {
		assert.Equal(t, fmt.Sprintf("entry %v", i+4), entry.Description)
	}
}

func Test_timeEntryUsecase_UpdateTimeEntry(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	err = TestTimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "updatedTimeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		ID:          id,
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}

	timeEntry.Description = "updatedTimeentry"
	err = TestTimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var notFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &notFoundError))

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entryList))
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfUserIdIsEmpty(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	timeEntry.UserId = uuid.Nil
	err = TestTimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfProjectIdIsEmpty(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	timeEntry.ProjectId = uuid.Nil
	err = TestTimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry.UserId = userId
	err = TestTimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var userNotFoundError *UserNotFoundError
	assert.True(t, errors.As(err, &userNotFoundError))

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry.ProjectId = projectId
	err = TestTimeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var projectNotFoundError *ProjectNotFoundError
	assert.True(t, errors.As(err, &projectNotFoundError))

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_DeleteTimeEntry(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	err = TestTimeEntryUsecase.DeleteTimeEntry(timeEntry.ID)
	assert.Nil(t, err)

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entryList))
}

func Test_timeEntryUsecase_DeleteTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := addUser(t, "user", "password", model.RoleList{model.RoleUser})
	project := addProject(t, "project", user)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := TestTimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = TestTimeEntryUsecase.DeleteTimeEntry(notExistingId)
	assert.NotNil(t, err)
	var notFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &notFoundError))

	entryList, err := TestTimeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
}

func assertTimesAreEqual(t *testing.T, time1 time.Time, time2 time.Time) {
	// We cannot check the milliseconds here because they get lost in the database:
	assert.Equal(t, time1.Hour(), time2.Hour())
	assert.Equal(t, time1.Minute(), time2.Minute())
	assert.Equal(t, time1.Second(), time2.Second())
}

func addTimeEntries(t *testing.T, count int, owner model.User, project model.Project) []model.TimeEntry {
	return addTimeEntriesWithStartIndex(t, 1, count, owner, project)
}

func addTimeEntriesWithStartIndex(t *testing.T, startIndex int, count int, owner model.User, project model.Project) []model.TimeEntry {
	var entries []model.TimeEntry
	startTime := time.Now()
	oneHour := 1000 * 1000 * 60 * 60 // duration is in nanoseconds
	for i := 0; i < count; i++ {
		entry := model.TimeEntry{
			Description: fmt.Sprintf("entry %v", startIndex+i),
			StartTime:   startTime.Add(time.Duration(oneHour * (count - i))),
			UserId:      owner.ID,
			ProjectId:   project.ID,
		}
		entries = append(entries, entry)
		err := TestTimeEntryUsecase.AddTimeEntry(&entry)
		assert.Nil(t, err)
	}
	return entries
}
