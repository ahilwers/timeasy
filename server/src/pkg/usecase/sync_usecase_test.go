package usecase

import (
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/stretchr/testify/assert"
)

func Test_syncUsecase_CanUpdatedEntriesBeFetchedWhenEntryIsNew(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	oldTimeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	oldTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	oldTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&oldTimeEntry)
	assert.Nil(t, err)

	newTimeEntry := model.TimeEntry{
		Description: "newTimeEntry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err = usecaseTest.TimeEntryUsecase.AddTimeEntry(&newTimeEntry)
	assert.Nil(t, err)

	changedEntries, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changedEntries))
	assert.Equal(t, "newTimeEntry", changedEntries[0].Description)
}

func Test_syncUsecase_CanUpdatedEntriesBeFetchedWhenEntryIsUpdated(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	oldTimeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	oldTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	oldTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&oldTimeEntry)
	assert.Nil(t, err)

	// The entry should not be returned now:
	changedEntries, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(changedEntries))

	//Update the timeentry:
	oldTimeEntry.Description = "updatedTimeEntry"
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&oldTimeEntry)
	assert.Nil(t, err)

	changedEntries, err = usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changedEntries))
	assert.Equal(t, "updatedTimeEntry", changedEntries[0].Description)
}

func Test_syncUsecase_CanUpdatedEntriesBeFetchedWhenEntryIsDeleted(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	oldTimeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	oldTimeEntry.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	oldTimeEntry.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&oldTimeEntry)
	assert.Nil(t, err)

	// The entry should not be returned now:
	changedEntries, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(changedEntries))

	//Delete the timeentry:
	err = usecaseTest.TimeEntryUsecase.DeleteTimeEntry(oldTimeEntry.ID)
	assert.Nil(t, err)

	changedEntries, err = usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changedEntries))
	assert.Equal(t, "timeentry", changedEntries[0].Description)
}

func Test_syncUsecase_CanUpdatedProjectsBeFetchedWhenEntryIsNew(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	oldProject := model.Project{
		Name:   "project",
		UserId: userId,
	}
	oldProject.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	oldProject.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err := usecaseTest.ProjectUsecase.AddProject(&oldProject)
	assert.Nil(t, err)

	newProject := model.Project{
		Name:   "newProject",
		UserId: userId,
	}
	err = usecaseTest.ProjectUsecase.AddProject(&newProject)
	assert.Nil(t, err)

	changedProjects, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changedProjects))
	assert.Equal(t, "newProject", changedProjects[0].Name)
}

func Test_syncUsecase_CanUpdatedProjectsBeFetchedWhenEntryIsUpdated(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	oldProject := model.Project{
		Name:   "project",
		UserId: userId,
	}
	oldProject.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	oldProject.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err := usecaseTest.ProjectUsecase.AddProject(&oldProject)
	assert.Nil(t, err)

	// The project should not be returned now:
	changedProjects, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(changedProjects))

	//Update the project:
	oldProject.Name = "updatedProject"
	err = usecaseTest.ProjectUsecase.UpdateProject(&oldProject)
	assert.Nil(t, err)

	changedProjects, err = usecaseTest.SyncUsecase.GetChangedProjects(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changedProjects))
	assert.Equal(t, "updatedProject", changedProjects[0].Name)
}

func Test_syncUsecase_CanUpdatedProjectsBeFetchedWhenEntryIsDeleted(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	oldProject := model.Project{
		Name:   "project",
		UserId: userId,
	}
	oldProject.UpdatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	oldProject.CreatedAt = time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC)
	err := usecaseTest.ProjectUsecase.AddProject(&oldProject)
	assert.Nil(t, err)

	// The project should not be returned now:
	changedProjects, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(changedProjects))

	//Delete the project:
	err = usecaseTest.ProjectUsecase.DeleteProject(oldProject.ID)
	assert.Nil(t, err)

	changedProjects, err = usecaseTest.SyncUsecase.GetChangedProjects(userId, time.Date(2023, 8, 31, 0, 0, 0, 0, time.UTC))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(changedProjects))
	assert.Equal(t, "project", changedProjects[0].Name)
}
