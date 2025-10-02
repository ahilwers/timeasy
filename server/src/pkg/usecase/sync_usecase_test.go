package usecase

import (
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
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

func Test_syncUsecase_MergesDuplicateOpenTimeEntries(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "test project", userId)
	clientId := GetTestClientId(t)

	// Create fixed times for testing
	earlierTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	laterTime := time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)

	// Create an existing open time entry in the database (simulating one from another device)
	existingOpenEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "existing open entry",
		StartTime:   earlierTime, // Earlier time
		EndTime:     time.Time{}, // Zero time means open
		UserId:      userId,
		ProjectId:   project.ID,
	}
	
	tx, err := usecaseTest.TimeEntryRepository.BeginTransaction()
	assert.Nil(t, err)
	err = usecaseTest.TimeEntryRepository.AddTimeEntry(&existingOpenEntry, tx)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

	// Create a new open time entry coming from sync (simulating one from current device)
	newOpenEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "new open entry",
		StartTime:   laterTime, // Later time
		EndTime:     time.Time{}, // Zero time means open
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Create sync data with the new open entry
	syncData := model.SyncData{
		TimeEntriesToBeCreated: []model.TimeEntry{newOpenEntry},
	}

	// Process the sync data - this should trigger the merge
	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify that we now have only one open time entry for this project
	openEntries, err := usecaseTest.TimeEntryRepository.GetOpenTimeEntriesForProject(userId, project.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(openEntries), "Should have exactly one open time entry after merge")

	// Verify the merged entry has the correct properties
	mergedEntry := openEntries[0]
	assert.Equal(t, newOpenEntry.ID, mergedEntry.ID, "Should use the ID from the incoming entry")
	assert.True(t, mergedEntry.StartTime.UTC().Equal(laterTime), "Should use the latest start time")
	assert.Contains(t, mergedEntry.Description, "existing open entry", "Should contain the existing entry's description")
	assert.Contains(t, mergedEntry.Description, "new open entry", "Should contain the new entry's description")
	assert.True(t, mergedEntry.EndTime.IsZero(), "Should still be an open entry")

	// Verify that changelog entries were created properly
	// The existing entry should be marked as deleted, and the new merged entry should be created
	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 1, 0, "")
	assert.Nil(t, err)
	
	// Should have one created entry (the merged one) and one deleted entry (the existing one)
	assert.Equal(t, 1, len(result.Created), "Should have one created entry")
	assert.Equal(t, 1, len(result.Deleted), "Should have one deleted entry")
	
	// The created entry should be the merged one
	assert.Equal(t, newOpenEntry.ID, result.Created[0].ID)
	assert.Contains(t, result.Created[0].Description, "existing open entry")
	assert.Contains(t, result.Created[0].Description, "new open entry")
	
	// The deleted entry should be the existing one
	assert.Equal(t, existingOpenEntry.ID, result.Deleted[0].ID)
}

func Test_syncUsecase_DoesNotMergeClosedTimeEntries(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "test project", userId)
	clientId := GetTestClientId(t)

	// Create a closed time entry coming from sync
	closedEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "closed entry",
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now().Add(-1 * time.Hour), // Has end time, so it's closed
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Create sync data with the closed entry
	syncData := model.SyncData{
		TimeEntriesToBeCreated: []model.TimeEntry{closedEntry},
	}

	// Process the sync data - this should NOT trigger merge logic
	err := usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the entry was created normally
	createdEntry, err := usecaseTest.SyncUsecase.GetTimeEntryById(closedEntry.ID)
	assert.Nil(t, err)
	assert.NotNil(t, createdEntry)
	assert.Equal(t, closedEntry.Description, createdEntry.Description)
	assert.False(t, createdEntry.EndTime.IsZero(), "Should have an end time")
}

func Test_syncUsecase_MergesMultipleOpenTimeEntries(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "test project", userId)
	clientId := GetTestClientId(t)

	// Create fixed times for testing
	earliestTime := time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)  // 9 AM
	middleTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)   // 10 AM
	latestTime := time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)   // 11 AM

	// Create two existing open time entries in the database
	firstOpenEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "first entry",
		StartTime:   earliestTime, // Earliest
		EndTime:     time.Time{}, // Open
		UserId:      userId,
		ProjectId:   project.ID,
	}
	
	secondOpenEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "second entry",
		StartTime:   middleTime,
		EndTime:     time.Time{}, // Open
		UserId:      userId,
		ProjectId:   project.ID,
	}
	
	tx, err := usecaseTest.TimeEntryRepository.BeginTransaction()
	assert.Nil(t, err)
	err = usecaseTest.TimeEntryRepository.AddTimeEntry(&firstOpenEntry, tx)
	assert.Nil(t, err)
	err = usecaseTest.TimeEntryRepository.AddTimeEntry(&secondOpenEntry, tx)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

	// Create a third open time entry coming from sync
	thirdOpenEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "third entry",
		StartTime:   latestTime, // Latest
		EndTime:     time.Time{}, // Open
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Create sync data
	syncData := model.SyncData{
		TimeEntriesToBeCreated: []model.TimeEntry{thirdOpenEntry},
	}

	// Process the sync data
	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify that we now have only one open time entry
	openEntries, err := usecaseTest.TimeEntryRepository.GetOpenTimeEntriesForProject(userId, project.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(openEntries), "Should have exactly one open time entry after merge")

	// Verify the merged entry properties
	mergedEntry := openEntries[0]
	assert.Equal(t, thirdOpenEntry.ID, mergedEntry.ID, "Should use the ID from the incoming entry")
	assert.True(t, mergedEntry.StartTime.UTC().Equal(latestTime), "Should use the latest start time")
	
	// Should contain all three descriptions
	assert.Contains(t, mergedEntry.Description, "first entry")
	assert.Contains(t, mergedEntry.Description, "second entry") 
	assert.Contains(t, mergedEntry.Description, "third entry")
}

func Test_syncUsecase_MobileAppReceivesMergedEntryBack(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "test project", userId)
	mobileClientId := "mobile-app-123"

	// Simulate scenario: Web app started timing at 10:00
	webStartTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	webEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "web entry",
		StartTime:   webStartTime,
		EndTime:     time.Time{}, // Open
		UserId:      userId,
		ProjectId:   project.ID,
	}
	
	// Add web entry directly to database (simulating it was already synced from web)
	tx, err := usecaseTest.TimeEntryRepository.BeginTransaction()
	assert.Nil(t, err)
	err = usecaseTest.TimeEntryRepository.AddTimeEntry(&webEntry, tx)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

	// Mobile app starts timing at 11:00 (offline, then syncs)
	mobileStartTime := time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC)
	mobileEntry := model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		Description: "mobile entry",
		StartTime:   mobileStartTime,
		EndTime:     time.Time{}, // Open
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Mobile app syncs its entry to server
	syncData := model.SyncData{
		TimeEntriesToBeCreated: []model.TimeEntry{mobileEntry},
	}

	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, mobileClientId)
	assert.Nil(t, err)

	// Get the latest changelog ID to simulate the mobile app's next sync
	latestChangelogId, err := usecaseTest.SyncUsecase.GetLatestChangelogEntryId()
	assert.Nil(t, err)

	// Mobile app requests changes from server (excluding its own changes)
	// This should include the merged entry because it was created by "server-merge", not the mobile client
	result, err := usecaseTest.SyncUsecase.GetChangedTimeEntries(userId, 1, latestChangelogId, mobileClientId)
	assert.Nil(t, err)

	// The mobile app should receive back the merged entry with updated start time
	assert.Equal(t, 1, len(result.Created), "Mobile app should receive the merged entry back")
	receivedEntry := result.Created[0]
	
	// Verify the received entry has the mobile app's ID and the later start time
	assert.Equal(t, mobileEntry.ID, receivedEntry.ID, "Should have mobile app's entry ID")
	assert.True(t, receivedEntry.StartTime.UTC().Equal(mobileStartTime), "Should have the later start time from mobile entry")
	assert.Contains(t, receivedEntry.Description, "web entry", "Should contain web entry description")
	assert.Contains(t, receivedEntry.Description, "mobile entry", "Should contain mobile entry description")
	
	// Mobile app should also receive deletion of the web entry
	assert.Equal(t, 1, len(result.Deleted), "Mobile app should receive deletion of web entry")
	assert.Equal(t, webEntry.ID, result.Deleted[0].ID, "Should delete the web entry")
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

func Test_syncUsecase_UpdateOfNonExistentEntryCreatesEntry(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)
	clientId := GetTestClientId(t)

	// Create a time entry that doesn't exist in the database (simulating a missed sync)
	nonExistentEntryId, err := uuid.NewV4()
	assert.Nil(t, err)

	entryToUpdate := model.TimeEntry{
		ID:          nonExistentEntryId,
		Description: "updated entry",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Try to sync an update to this non-existent entry
	syncData := model.SyncData{
		TimeEntriesToBeUpdated: []model.TimeEntry{entryToUpdate},
	}

	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the entry was created (not updated)
	createdEntry, err := usecaseTest.TimeEntryUsecase.GetTimeEntryById(nonExistentEntryId)
	assert.Nil(t, err)
	assert.NotNil(t, createdEntry)
	assert.Equal(t, "updated entry", createdEntry.Description)
	assert.Equal(t, userId, createdEntry.UserId)
	assert.Equal(t, project.ID, createdEntry.ProjectId)
}

func Test_syncUsecase_DeletionOfNonExistentEntrySucceeds(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)
	clientId := GetTestClientId(t)

	// Create a time entry that doesn't exist in the database (simulating a missed sync or race condition)
	nonExistentEntryId, err := uuid.NewV4()
	assert.Nil(t, err)

	entryToDelete := model.TimeEntry{
		ID:          nonExistentEntryId,
		Description: "entry to delete",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Try to sync a deletion of this non-existent entry
	syncData := model.SyncData{
		TimeEntriesToBeDeleted: []model.TimeEntry{entryToDelete},
	}

	// This should succeed (not fail) even though the entry doesn't exist
	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the entry still doesn't exist (since it was never there to begin with)
	deletedEntry, err := usecaseTest.TimeEntryUsecase.GetTimeEntryById(nonExistentEntryId)
	assert.NotNil(t, err) // Should get an error because entry doesn't exist
	assert.Nil(t, deletedEntry)
}

func Test_syncUsecase_MixedDeletionWithExistingAndNonExistentEntries(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, usecaseTest.ProjectUsecase, "project", userId)
	clientId := GetTestClientId(t)

	// Create an actual time entry in the database
	existingEntry := model.TimeEntry{
		Description: "existing entry",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&existingEntry, userId, clientId)
	assert.Nil(t, err)

	// Create a non-existent entry
	nonExistentEntryId, err := uuid.NewV4()
	assert.Nil(t, err)
	nonExistentEntry := model.TimeEntry{
		ID:          nonExistentEntryId,
		Description: "non-existent entry",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserId:      userId,
		ProjectId:   project.ID,
	}

	// Try to delete both entries (one exists, one doesn't)
	syncData := model.SyncData{
		TimeEntriesToBeDeleted: []model.TimeEntry{existingEntry, nonExistentEntry},
	}

	// This should succeed overall, even with the non-existent entry
	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the existing entry was actually deleted
	deletedEntry, err := usecaseTest.TimeEntryUsecase.GetTimeEntryById(existingEntry.ID)
	assert.NotNil(t, err) // Should get an error because entry was deleted
	assert.Nil(t, deletedEntry)

	// Verify the non-existent entry still doesn't exist
	stillNonExistent, err := usecaseTest.TimeEntryUsecase.GetTimeEntryById(nonExistentEntryId)
	assert.NotNil(t, err) // Should get an error because entry never existed
	assert.Nil(t, stillNonExistent)
}

func Test_syncUsecase_UpdateOfNonExistentProjectCreatesProject(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create a project that doesn't exist in the database (simulating a missed sync)
	nonExistentProjectId, err := uuid.NewV4()
	assert.Nil(t, err)

	projectToUpdate := model.Project{
		ID:     nonExistentProjectId,
		Name:   "updated project",
		UserId: userId,
		Color:  "#FF0000",
	}

	// Try to sync an update to this non-existent project
	syncData := model.SyncData{
		ProjectsToBeUpdated: []model.Project{projectToUpdate},
	}

	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the project was created (not updated)
	createdProject, err := usecaseTest.ProjectUsecase.GetProjectById(nonExistentProjectId)
	assert.Nil(t, err)
	assert.NotNil(t, createdProject)
	assert.Equal(t, "updated project", createdProject.Name)
	assert.Equal(t, userId, createdProject.UserId)
	assert.Equal(t, "#FF0000", createdProject.Color)
}

func Test_syncUsecase_DeletionOfNonExistentProjectSucceeds(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create a project that doesn't exist in the database (simulating a missed sync or race condition)
	nonExistentProjectId, err := uuid.NewV4()
	assert.Nil(t, err)

	projectToDelete := model.Project{
		ID:     nonExistentProjectId,
		Name:   "project to delete",
		UserId: userId,
	}

	// Try to sync a deletion of this non-existent project
	syncData := model.SyncData{
		ProjectsToBeDeleted: []model.Project{projectToDelete},
	}

	// This should succeed (not fail) even though the project doesn't exist
	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the project still doesn't exist (since it was never there to begin with)
	deletedProject, err := usecaseTest.ProjectUsecase.GetProjectById(nonExistentProjectId)
	assert.NotNil(t, err) // Should get an error because project doesn't exist
	assert.Nil(t, deletedProject)
}

func Test_syncUsecase_MixedProjectDeletionWithExistingAndNonExistentProjects(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create an actual project in the database
	existingProject := model.Project{
		Name:   "existing project",
		UserId: userId,
	}
	err := usecaseTest.ProjectUsecase.AddProject(&existingProject, userId, clientId)
	assert.Nil(t, err)

	// Create a non-existent project
	nonExistentProjectId, err := uuid.NewV4()
	assert.Nil(t, err)
	nonExistentProject := model.Project{
		ID:     nonExistentProjectId,
		Name:   "non-existent project",
		UserId: userId,
	}

	// Try to delete both projects (one exists, one doesn't)
	syncData := model.SyncData{
		ProjectsToBeDeleted: []model.Project{existingProject, nonExistentProject},
	}

	// This should succeed overall, even with the non-existent project
	err = usecaseTest.SyncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	assert.Nil(t, err)

	// Verify the existing project was actually deleted
	deletedProject, err := usecaseTest.ProjectUsecase.GetProjectById(existingProject.ID)
	assert.NotNil(t, err) // Should get an error because project was deleted
	assert.Nil(t, deletedProject)

	// Verify the non-existent project still doesn't exist
	stillNonExistent, err := usecaseTest.ProjectUsecase.GetProjectById(nonExistentProjectId)
	assert.NotNil(t, err) // Should get an error because project never existed
	assert.Nil(t, stillNonExistent)
}
