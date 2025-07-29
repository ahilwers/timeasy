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
	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 2, 0, "")
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

	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 0, 0, "")
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

	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 2, 0, "")
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

	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 2, 0, "")
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

	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, 1, 0, "")
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

	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, 1, 0, "")
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

	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, 1, 0, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 1, len(result.Deleted))
	assert.Equal(t, "project", result.Deleted[0].Name)
}

func Test_syncUsecase_CanFetchEntriesWithinSpecificChangelogIdRange(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create first project
	firstProject := model.Project{
		Name:   "first project",
		UserId: userId,
		Color:  "#1E90FF", // Blue
	}
	err := usecaseTest.ProjectUsecase.AddProject(&firstProject, userId, clientId)
	assert.Nil(t, err)

	// Get the latest changelog entry ID after first project creation
	latestId, err := usecaseTest.SyncUsecase.GetLatestChangelogEntryId()
	assert.Nil(t, err)
	firstProjectChangelogId := latestId

	// Create second project
	secondProject := model.Project{
		Name:   "second project",
		UserId: userId,
		Color:  "#32CD32", // Green
	}
	err = usecaseTest.ProjectUsecase.AddProject(&secondProject, userId, clientId)
	assert.Nil(t, err)

	// Get the latest changelog entry ID after second project creation
	latestId, err = usecaseTest.SyncUsecase.GetLatestChangelogEntryId()
	assert.Nil(t, err)
	secondProjectChangelogId := latestId

	// Create third project
	thirdProject := model.Project{
		Name:   "third project",
		UserId: userId,
		Color:  "#FF6347", // Red
	}
	err = usecaseTest.ProjectUsecase.AddProject(&thirdProject, userId, clientId)
	assert.Nil(t, err)

	// Get the latest changelog entry ID after third project creation
	latestId, err = usecaseTest.SyncUsecase.GetLatestChangelogEntryId()
	assert.Nil(t, err)
	thirdProjectChangelogId := latestId

	// Test fetching only the second project (between first and third)
	result, err := usecaseTest.SyncUsecase.GetChangedProjects(userId, firstProjectChangelogId, secondProjectChangelogId, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	assert.Equal(t, "second project", result.Created[0].Name)

	// Test fetching the second and third projects
	result, err = usecaseTest.SyncUsecase.GetChangedProjects(userId, firstProjectChangelogId, thirdProjectChangelogId, "")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result.Created))
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
	// The results should be ordered by changelog ID
	projectNames := []string{result.Created[0].Name, result.Created[1].Name}
	assert.Contains(t, projectNames, "second project")
	assert.Contains(t, projectNames, "third project")
}

func Test_syncUsecase_CanFetchTimeEntriesWithinSpecificChangelogIdRange(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create a project for the time entries
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)

	// Create first time entry
	firstTimeEntry := model.TimeEntry{
		Description: "first time entry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&firstTimeEntry, userId, clientId)
	assert.Nil(t, err)

	// Get the latest changelog entry ID after first time entry creation
	firstTimeEntryChangelogId, err := usecaseTest.SyncUsecase.GetLatestChangelogEntryId()
	assert.Nil(t, err)

	// Create second time entry
	secondTimeEntry := model.TimeEntry{
		Description: "second time entry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err = usecaseTest.TimeEntryUsecase.AddTimeEntry(&secondTimeEntry, userId, clientId)
	assert.Nil(t, err)

	// Get the latest changelog entry ID after second time entry creation
	secondTimeEntryChangelogId, err := usecaseTest.SyncUsecase.GetLatestChangelogEntryId()
	assert.Nil(t, err)

	// Create third time entry
	thirdTimeEntry := model.TimeEntry{
		Description: "third time entry",
		StartTime:   time.Now(),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err = usecaseTest.TimeEntryUsecase.AddTimeEntry(&thirdTimeEntry, userId, clientId)
	assert.Nil(t, err)

	// Test fetching only the second time entry (between firstTimeEntryChangelogId and secondTimeEntryChangelogId)
	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, firstTimeEntryChangelogId, secondTimeEntryChangelogId, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Created))
	assert.Equal(t, "second time entry", result.Created[0].Description)
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))

	// Test fetching time entries after secondTimeEntryChangelogId
	result, err = usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, secondTimeEntryChangelogId, 0, "")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Created))
	assert.Equal(t, "third time entry", result.Created[0].Description)
	assert.Equal(t, 0, len(result.Updated))
	assert.Equal(t, 0, len(result.Deleted))
}
