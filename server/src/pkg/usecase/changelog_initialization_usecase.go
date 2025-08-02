package usecase

import (
	"log"
	"time"

	"github.com/gofrs/uuid"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
)

type ChangelogInitializationUsecase interface {
	InitializeChangelog() error
}

type changelogInitializationUsecase struct {
	changelogRepo   repository.ChangelogRepository
	projectRepo     repository.ProjectRepository
	timeEntryRepo   repository.TimeEntryRepository
	teamRepo        repository.TeamRepository
}

func NewChangelogInitializationUsecase(
	changelogRepo repository.ChangelogRepository,
	projectRepo repository.ProjectRepository,
	timeEntryRepo repository.TimeEntryRepository,
	teamRepo repository.TeamRepository,
) ChangelogInitializationUsecase {
	return &changelogInitializationUsecase{
		changelogRepo: changelogRepo,
		projectRepo:   projectRepo,
		timeEntryRepo: timeEntryRepo,
		teamRepo:      teamRepo,
	}
}

func (u *changelogInitializationUsecase) InitializeChangelog() error {
	log.Println("Checking if changelog initialization is needed...")

	hasEntries, err := u.changelogRepo.HasAnyEntries()
	if err != nil {
		return err
	}

	if hasEntries {
		log.Println("Changelog already has entries, skipping initialization")
		return nil
	}

	log.Println("Changelog is empty, checking for existing data to initialize...")

	tx, err := u.changelogRepo.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	initTime := time.Now().UTC()
	entriesCreated := 0

	// Initialize projects
	projects, err := u.projectRepo.GetAllProjects()
	if err != nil {
		return err
	}

	for _, project := range projects {
		entry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeProject,
			EntityID:        project.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   project.UserId,
			ChangedByClient: "", // Empty as specified
			ChangedAt:       initTime,
		}
		if err := u.changelogRepo.AddChangelogEntry(entry, tx); err != nil {
			return err
		}
		entriesCreated++
	}

	// Initialize time entries
	timeEntries, err := u.timeEntryRepo.GetAllTimeEntries()
	if err != nil {
		return err
	}

	for _, timeEntry := range timeEntries {
		entry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeTimeEntry,
			EntityID:        timeEntry.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   timeEntry.UserId,
			ChangedByClient: "", // Empty as specified
			ChangedAt:       initTime,
		}
		if err := u.changelogRepo.AddChangelogEntry(entry, tx); err != nil {
			return err
		}
		entriesCreated++
	}

	// Initialize teams
	teams, err := u.teamRepo.GetAllTeams()
	if err != nil {
		return err
	}

	for _, team := range teams {
		// For teams, we need to get the user who created it
		// Since we don't have that info, we'll use a nil UUID and handle it appropriately
		entry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeTeam,
			EntityID:        team.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   uuid.Nil, // We don't have the creator info for teams
			ChangedByClient: "", // Empty as specified
			ChangedAt:       initTime,
		}
		if err := u.changelogRepo.AddChangelogEntry(entry, tx); err != nil {
			return err
		}
		entriesCreated++
	}

	// Initialize user team assignments
	assignments, err := u.teamRepo.GetAllUserTeamAssignments()
	if err != nil {
		return err
	}

	for _, assignment := range assignments {
		entry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeUserTeamAssignment,
			EntityID:        assignment.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   assignment.UserID,
			ChangedByClient: "", // Empty as specified
			ChangedAt:       initTime,
		}
		if err := u.changelogRepo.AddChangelogEntry(entry, tx); err != nil {
			return err
		}
		entriesCreated++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Changelog initialization completed. Created %d changelog entries", entriesCreated)
	return nil
}