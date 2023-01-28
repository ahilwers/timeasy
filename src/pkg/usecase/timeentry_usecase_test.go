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

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
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

func Test_timeEntryUsecase_AddTimeEntryFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)

	userId, err := uuid.NewV4()
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err = timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var userNotFoundError *UserNotFoundError
	assert.True(t, errors.As(err, &userNotFoundError))
}

func Test_timeEntryUsecase_AddTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   projectId,
	}
	err = timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var projectNotFoundError *ProjectNotFoundError
	assert.True(t, errors.As(err, &projectNotFoundError))
}

func Test_timeEntryUsecase_GetTimeEntryById(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
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

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)

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

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})
	otherUser := addUser(t, userUsecase, "otheruser", "otherpassword", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)

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

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)
	otherProject := addProject(t, projectUsecase, "otherproject", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)

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

func Test_timeEntryUsecase_UpdateTimeEntry(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	err = timeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.Nil(t, err)

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "updatedTimeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
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
	err = timeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var notFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &notFoundError))

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entryList))
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfUserIdIsEmpty(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	timeEntry.UserId = uuid.Nil
	err = timeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfProjectIdIsEmpty(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	timeEntry.ProjectId = uuid.Nil
	err = timeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var entityIncompleteError *EntityIncompleteError
	assert.True(t, errors.As(err, &entityIncompleteError))

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry.UserId = userId
	err = timeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var userNotFoundError *UserNotFoundError
	assert.True(t, errors.As(err, &userNotFoundError))

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_UpdateTimeEntryFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	timeEntry.Description = "updatedTimeentry"
	projectId, err := uuid.NewV4()
	assert.Nil(t, err)
	timeEntry.ProjectId = projectId
	err = timeEntryUsecase.UpdateTimeEntry(&timeEntry)
	assert.NotNil(t, err)
	var projectNotFoundError *ProjectNotFoundError
	assert.True(t, errors.As(err, &projectNotFoundError))

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
	assert.Equal(t, "timeentry", entryList[0].Description)
	assertTimesAreEqual(t, timeEntry.StartTime, entryList[0].StartTime)
	assert.Equal(t, user.ID, entryList[0].UserId)
	assert.Equal(t, project.ID, entryList[0].ProjectId)
	assert.True(t, entryList[0].EndTime.IsZero())
}

func Test_timeEntryUsecase_DeleteTimeEntry(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	err = timeEntryUsecase.DeleteTimeEntry(timeEntry.ID)
	assert.Nil(t, err)

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(entryList))
}

func Test_timeEntryUsecase_DeleteTimeEntryFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := addUser(t, userUsecase, "user", "password", model.RoleList{model.RoleUser})

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := NewProjectUsecase(projectRepo)
	project := addProject(t, projectUsecase, "project", user)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	timeEntryUsecase := NewTimeEntryUsecase(timeEntryRepo, userUsecase, projectUsecase)
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      user.ID,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = timeEntryUsecase.DeleteTimeEntry(notExistingId)
	assert.NotNil(t, err)
	var notFoundError *EntityNotFoundError
	assert.True(t, errors.As(err, &notFoundError))

	entryList, err := timeEntryUsecase.GetAllTimeEntriesOfUser(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(entryList))
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
