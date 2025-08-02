package usecase

import (
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"timeasy-server/pkg/domain/model"
)

func Test_changelogInitializationUsecase_InitializeChangelog(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)

	// Create some data directly in the repositories (bypassing changelog creation)
	// Add a team directly
	team := model.Team{Name1: "Test Team"}
	err := usecaseTest.TeamRepository.AddTeam(&team)
	assert.Nil(t, err)

	// Add a user team assignment directly
	assignment := model.UserTeamAssignment{
		UserID: userId,
		TeamID: team.ID,
		Roles:  []string{"user"},
	}
	err = usecaseTest.TeamRepository.AddUserTeamAssignment(&assignment)
	assert.Nil(t, err)

	// Add a project directly
	project := model.Project{Name: "Test Project", UserId: userId}
	tx, err := usecaseTest.ProjectRepository.BeginTransaction()
	assert.Nil(t, err)
	err = usecaseTest.ProjectRepository.AddProject(&project, tx)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

	// Add a time entry directly
	timeEntry := model.TimeEntry{
		Description: "Test time entry",
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now().Add(-1 * time.Hour),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	tx, err = usecaseTest.TimeEntryRepository.BeginTransaction()
	assert.Nil(t, err)
	err = usecaseTest.TimeEntryRepository.AddTimeEntry(&timeEntry, tx)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

	// Verify changelog is initially empty
	hasEntries, err := usecaseTest.ChangelogRepository.HasAnyEntries()
	assert.Nil(t, err)
	assert.False(t, hasEntries)

	// Create the changelog initialization usecase
	changelogInitUsecase := NewChangelogInitializationUsecase(
		usecaseTest.ChangelogRepository,
		usecaseTest.ProjectRepository,
		usecaseTest.TimeEntryRepository,
		usecaseTest.TeamRepository,
	)

	// Run changelog initialization
	err = changelogInitUsecase.InitializeChangelog()
	assert.Nil(t, err)

	// Verify changelog now has entries
	hasEntries, err = usecaseTest.ChangelogRepository.HasAnyEntries()
	assert.Nil(t, err)
	assert.True(t, hasEntries)

	// Get all changelog entries and verify they are correct
	entries, err := usecaseTest.ChangelogRepository.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.True(t, len(entries) >= 4) // project, time entry, team, and user team assignment

	// Verify we have entries for each entity type
	foundProject := false
	foundTimeEntry := false
	foundTeam := false
	foundUserTeamAssignment := false

	for _, entry := range entries {
		assert.Equal(t, model.OperationCreated, entry.Operation)
		assert.Equal(t, "", entry.ChangedByClient) // Should be empty as specified

		switch entry.EntityType {
		case model.EntityTypeProject:
			assert.Equal(t, project.ID, entry.EntityID)
			assert.Equal(t, userId, entry.ChangedByUser)
			foundProject = true
		case model.EntityTypeTimeEntry:
			assert.Equal(t, timeEntry.ID, entry.EntityID)
			assert.Equal(t, userId, entry.ChangedByUser)
			foundTimeEntry = true
		case model.EntityTypeTeam:
			assert.Equal(t, team.ID, entry.EntityID)
			foundTeam = true
		case model.EntityTypeUserTeamAssignment:
			assert.Equal(t, assignment.ID, entry.EntityID)
			assert.Equal(t, userId, entry.ChangedByUser)
			foundUserTeamAssignment = true
		}
	}

	assert.True(t, foundProject, "Should have created changelog entry for project")
	assert.True(t, foundTimeEntry, "Should have created changelog entry for time entry")
	assert.True(t, foundTeam, "Should have created changelog entry for team")
	assert.True(t, foundUserTeamAssignment, "Should have created changelog entry for user team assignment")
}

func Test_changelogInitializationUsecase_InitializeChangelogSkipsIfNotEmpty(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	// Create the changelog initialization usecase
	changelogInitUsecase := NewChangelogInitializationUsecase(
		usecaseTest.ChangelogRepository,
		usecaseTest.ProjectRepository,
		usecaseTest.TimeEntryRepository,
		usecaseTest.TeamRepository,
	)

	// Add some test data
	userId := GetTestUserId(t)

	// Add a project which will create a changelog entry
	project := addProject(t, usecaseTest.ProjectUsecase, "Test Project", userId)

	// Verify changelog has one entry
	entries, err := usecaseTest.ChangelogRepository.GetChangelogEntries(nil)
	assert.Nil(t, err)
	initialCount := len(entries)
	assert.True(t, initialCount > 0)

	// Run changelog initialization - it should skip because changelog is not empty
	err = changelogInitUsecase.InitializeChangelog()
	assert.Nil(t, err)

	// Verify no additional entries were created
	entries, err = usecaseTest.ChangelogRepository.GetChangelogEntries(nil)
	assert.Nil(t, err)
	assert.Equal(t, initialCount, len(entries))

	// Verify the existing entry is for the project we added
	found := false
	for _, entry := range entries {
		if entry.EntityType == model.EntityTypeProject && entry.EntityID == project.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Should still have the original project changelog entry")
}

// Helper function to add a single time entry for testing
func addSingleTimeEntry(t *testing.T, timeEntryUsecase TimeEntryUsecase, project model.Project, userId uuid.UUID) model.TimeEntry {
	clientId := GetTestClientId(t)
	entry := model.TimeEntry{
		Description: "Test time entry",
		StartTime:   time.Now().Add(-2 * time.Hour),
		EndTime:     time.Now().Add(-1 * time.Hour),
		UserId:      userId,
		ProjectId:   project.ID,
	}
	err := timeEntryUsecase.AddTimeEntry(&entry, userId, clientId)
	assert.Nil(t, err)
	return entry
}