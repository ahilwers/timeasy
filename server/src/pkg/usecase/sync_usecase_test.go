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
	clientId := GetTestClientId(t)

	oldTimeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&oldTimeEntry, userId, clientId)
	assert.Nil(t, err)

	newTimeEntry := model.TimeEntry{
		Description: "newTimeEntry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err = usecaseTest.TimeEntryUsecase.AddTimeEntry(&newTimeEntry, userId, clientId)
	assert.Nil(t, err)

	// To get the correct entry we need to use "2" here because "1" is the changelog entry for the project, "2" is the first timeentry and "3" is the entry we acutally want:
	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 2, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	assert.Equal(t, "newTimeEntry", result.Created[0].Description)
}

func Test_syncUsecase_IsOnlyTheLatestUpdateOfanEntryReturned(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)
	clientId := GetTestClientId(t)

	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry, userId, clientId)
	assert.Nil(t, err)

	timeEntry.Description = "UpdatedTimeEntry"
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&timeEntry, userId, clientId)
	assert.Nil(t, err)

	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 0, "")
	assert.Nil(t, err)
	// Since we're starting from 0, we should see this as a created entry (not updated)
	// because the latest operation for this entry is what matters
	assert.Equal(t, 1, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	assert.Equal(t, "UpdatedTimeEntry", result.Created[0].Description)
}

func Test_syncUsecase_CanUpdatedEntriesBeFetchedWhenEntryIsUpdated(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)
	clientId := GetTestClientId(t)

	oldTimeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&oldTimeEntry, userId, clientId)
	assert.Nil(t, err)

	//Update the timeentry:
	oldTimeEntry.Description = "updatedTimeEntry"
	err = usecaseTest.TimeEntryUsecase.UpdateTimeEntry(&oldTimeEntry, userId, clientId)
	assert.Nil(t, err)

	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 2, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result.Created))
	assert.Equal(t, 1, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	assert.Equal(t, "updatedTimeEntry", result.Updated[0].Description)
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
	clientId := GetTestClientId(t)
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&oldTimeEntry, userId, clientId)
	assert.Nil(t, err)

	//Delete the timeentry:
	err = usecaseTest.TimeEntryUsecase.DeleteTimeEntry(oldTimeEntry.ID, userId, clientId)
	assert.Nil(t, err)

	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 2, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 1, len(result.Deleted))
	assert.Equal(t, "timeentry", result.Deleted[0].Description)
}

func Test_syncUsecase_CanUpdatedProjectsBeFetchedWhenEntryIsNew(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	oldProject := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err := usecaseTest.ProjectUsecase.AddProject(&oldProject, userId, clientId)
	assert.Nil(t, err)

	newProject := model.Project{
		Name:   "newProject",
		UserId: userId,
	}
	err = usecaseTest.ProjectUsecase.AddProject(&newProject, userId, clientId)
	assert.Nil(t, err)

	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, 1, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	assert.Equal(t, "newProject", result.Created[0].Name)
}

func Test_syncUsecase_CanUpdatedProjectsBeFetchedWhenEntryIsUpdated(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	oldProject := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err := usecaseTest.ProjectUsecase.AddProject(&oldProject, userId, clientId)
	assert.Nil(t, err)

	//Update the project:
	oldProject.Name = "updatedProject"
	err = usecaseTest.ProjectUsecase.UpdateProject(&oldProject, userId, clientId)
	assert.Nil(t, err)

	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, 1, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result.Created))
	assert.Equal(t, 1, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	assert.Equal(t, "updatedProject", result.Updated[0].Name)
}

func Test_syncUsecase_CanUpdatedProjectsBeFetchedWhenEntryIsDeleted(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	oldProject := model.Project{
		Name:   "project",
		UserId: userId,
	}
	err := usecaseTest.ProjectUsecase.AddProject(&oldProject, userId, clientId)
	assert.Nil(t, err)

	// The project should not be returned now:
	//Delete the project:
	err = usecaseTest.ProjectUsecase.DeleteProject(oldProject.ID, userId, clientId)
	assert.Nil(t, err)

	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, 1, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 1, len(result.Deleted))
	assert.Equal(t, "project", result.Deleted[0].Name)
}
