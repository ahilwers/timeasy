package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type SyncUsecase interface {
	UpdateAndDeleteData(data model.SyncData, userId uuid.UUID, clientId string) error
	GetChangedTimeEntries(userId uuid.UUID, sinceTimeLogEntry int64, untilTimeLogEntry int64, excludeClientId string) (model.TimeEntrySyncResult, error)
	GetChangedProjects(userId uuid.UUID, sinceTimeLogEntry int64, untilTimeLogEntry int64, excludeClientId string) (model.ProjectSyncResult, error)
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
	GetLatestChangelogEntryId() (int64, error)
}

type syncUsecase struct {
	syncRepository      repository.SyncRepository
	changelogRepository repository.ChangelogRepository
	projectRepository   repository.ProjectRepository
	timeEntryRepository repository.TimeEntryRepository
	externalUsecase     *ExternalIntegrationUseCase
}

func NewSyncUsecase(
	syncRepository repository.SyncRepository,
	changelogRepository repository.ChangelogRepository,
	projectRepository repository.ProjectRepository,
	timeEntryRepository repository.TimeEntryRepository,
	externalUsecase *ExternalIntegrationUseCase,
) SyncUsecase {
	return &syncUsecase{
		syncRepository:      syncRepository,
		changelogRepository: changelogRepository,
		projectRepository:   projectRepository,
		timeEntryRepository: timeEntryRepository,
		externalUsecase:     externalUsecase,
	}
}

// UpdateAndDeleteData processes the sync data by creating, updating, and deleting entries
// and adds appropriate changelog entries for each operation
func (usecase *syncUsecase) UpdateAndDeleteData(data model.SyncData, userId uuid.UUID, clientId string) error {
	tx, err := usecase.timeEntryRepository.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	err = usecase.processProjects(data, userId, clientId, tx)
	if err != nil {
		return err
	}

	err = usecase.processTimeEntries(data, userId, clientId, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (usecase *syncUsecase) processProjects(data model.SyncData, userId uuid.UUID, clientId string, tx model.Transaction) error {
	// Process projects to be created
	err := usecase.processProjectCreations(data.ProjectsToBeCreated, userId, clientId, tx)
	if err != nil {
		return err
	}

	err = usecase.processProjectUpdates(data.ProjectsToBeUpdated, userId, clientId, tx)
	if err != nil {
		return err
	}

	err = usecase.processProjectDeletions(data.ProjectsToBeDeleted, userId, clientId, tx)
	if err != nil {
		return err
	}

	return nil
}

func (usecase *syncUsecase) processProjectCreations(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Processing project creations", "count", len(projects), "user_id", userId, "client_id", clientId, "method", "sync")
	for i := range projects {
		project := &projects[i]

		// Check if project already exists (including deleted ones)
		existingProject, err := usecase.syncRepository.GetProjectById(project.ID)
		if err != nil && !errors.Is(err, repository.ErrEntityNotFound) {
			return err
		}

		if existingProject != nil {
			// Project exists, update it (this will resurrect if deleted)
			slog.Debug("Project creation found existing project, updating instead",
				"project_id", project.ID,
				"user_id", userId,
				"client_id", clientId,
				"method", "sync",
			)
			err = usecase.projectRepository.UpdateProject(project, tx)
			if err != nil {
				return err
			}
		} else {
			// Project doesn't exist, create it
			err = usecase.projectRepository.AddProject(project, tx)
			if err != nil {
				return err
			}
		}

		changelogEntry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeProject,
			EntityID:        project.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   userId,
			ChangedByClient: clientId,
		}
		err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (usecase *syncUsecase) processProjectUpdates(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Processing project updates", "count", len(projects), "user_id", userId, "client_id", clientId, "method", "sync")
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.UpdateProject(project, tx)
		if err != nil {
			// If entity not found, try to add it instead
			if errors.Is(err, repository.ErrEntityNotFound) {
				slog.Warn("Project to be updated not found, adding as new",
					"project_id", project.ID,
					"user_id", userId,
					"project_name", project.Name,
					"client_id", clientId,
					"method", "sync",
				)
				err = usecase.projectRepository.AddProject(project, tx)
				if err != nil {
					return err
				}

				changelogEntry := &model.ChangelogEntry{
					EntityType:      model.EntityTypeProject,
					EntityID:        project.ID,
					Operation:       model.OperationCreated,
					ChangedByUser:   userId,
					ChangedByClient: clientId,
				}
				err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			changelogEntry := &model.ChangelogEntry{
				EntityType:      model.EntityTypeProject,
				EntityID:        project.ID,
				Operation:       model.OperationUpdated,
				ChangedByUser:   userId,
				ChangedByClient: clientId,
			}
			err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (usecase *syncUsecase) processProjectDeletions(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Processing project deletions", "count", len(projects), "user_id", userId, "client_id", clientId, "method", "sync")
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.DeleteProject(project, tx)
		if err != nil {
			if errors.Is(err, repository.ErrEntityNotFound) {
				slog.Warn("Project to be deleted not found, skipping",
					"project_id", project.ID,
					"user_id", userId,
					"project_name", project.Name,
					"client_id", clientId,
					"method", "sync",
				)
				continue
			} else {
				return err
			}
		}

		changelogEntry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeProject,
			EntityID:        project.ID,
			Operation:       model.OperationDeleted,
			ChangedByUser:   userId,
			ChangedByClient: clientId,
		}
		err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (usecase *syncUsecase) processTimeEntries(data model.SyncData, userId uuid.UUID, clientId string, tx model.Transaction) error {
	err := usecase.processTimeEntryCreations(data.TimeEntriesToBeCreated, userId, clientId, tx)
	if err != nil {
		return err
	}
	err = usecase.processTimeEntryUpdates(data.TimeEntriesToBeUpdated, userId, clientId, tx)
	if err != nil {
		return err
	}
	err = usecase.processTimeEntryDeletions(data.TimeEntriesToBeDeleted, userId, clientId, tx)
	if err != nil {
		return err
	}
	return nil
}

func (usecase *syncUsecase) processTimeEntryCreations(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Processing time entry creations", "count", len(timeEntries), "user_id", userId, "client_id", clientId, "method", "sync")
	for i := range timeEntries {
		timeEntry := &timeEntries[i]

		// Process issue detection and resolution before saving
		err := usecase.processIssueDetection(context.Background(), timeEntry, userId)
		if err != nil {
			// Log error but don't fail the sync - issue detection is not critical
			slog.Warn("Failed to process issue detection during sync operation",
				"time_entry_id", timeEntry.ID,
				"user_id", userId,
				"project_id", timeEntry.ProjectId,
				"description", timeEntry.Description,
				"client_id", clientId,
				"error", err)
		}

		// Check if an entry with this ID already exists (including deleted ones)
		existingEntry, err := usecase.syncRepository.GetTimeEntryById(timeEntry.ID)
		if err != nil {
			return err
		}

		if existingEntry != nil {
			// Entry exists - treat this as an update instead of create
			slog.Debug("Time entry with ID already exists, treating CREATE as UPDATE",
				"time_entry_id", timeEntry.ID,
				"existing_deleted", existingEntry.Deleted,
				"incoming_deleted", timeEntry.Deleted,
				"user_id", userId,
				"client_id", clientId,
				"method", "sync")

			// Update the existing entry with data from the incoming entry
			err = usecase.timeEntryRepository.UpdateTimeEntry(timeEntry, tx)
			if err != nil {
				return err
			}

			changelogEntry := &model.ChangelogEntry{
				EntityType:      model.EntityTypeTimeEntry,
				EntityID:        timeEntry.ID,
				Operation:       model.OperationUpdated,
				ChangedByUser:   userId,
				ChangedByClient: clientId,
			}
			err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
			if err != nil {
				return err
			}
		} else {
			// No existing entry - proceed with normal creation flow
			// Check if this is an open time entry (no end time)
			if timeEntry.EndTime.IsZero() {
				err := usecase.mergeOpenTimeEntriesIfNeeded(timeEntry, userId, clientId, tx)
				if err != nil {
					return err
				}
			} else {
				err := usecase.timeEntryRepository.AddTimeEntry(timeEntry, tx)
				if err != nil {
					return err
				}

				changelogEntry := &model.ChangelogEntry{
					EntityType:      model.EntityTypeTimeEntry,
					EntityID:        timeEntry.ID,
					Operation:       model.OperationCreated,
					ChangedByUser:   userId,
					ChangedByClient: clientId,
				}
				err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (usecase *syncUsecase) processTimeEntryUpdates(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Processing time entry updates", "count", len(timeEntries), "user_id", userId, "client_id", clientId, "method", "sync")
	for i := range timeEntries {
		timeEntry := &timeEntries[i]

		// Process issue detection and resolution before saving
		err := usecase.processIssueDetection(context.Background(), timeEntry, userId)
		if err != nil {
			// Log error but don't fail the sync - issue detection is not critical
			slog.Warn("Failed to process issue detection during sync operation",
				"time_entry_id", timeEntry.ID,
				"user_id", userId,
				"project_id", timeEntry.ProjectId,
				"description", timeEntry.Description,
				"client_id", clientId,
				"error", err)
		}

		err = usecase.timeEntryRepository.UpdateTimeEntry(timeEntry, tx)
		if err != nil {
			// If entity not found, try to add it instead
			if errors.Is(err, repository.ErrEntityNotFound) {
				slog.Warn("Time entry to be updated not found, adding as new",
					"time_entry_id", timeEntry.ID,
					"user_id", userId,
					"project_id", timeEntry.ProjectId,
					"description", timeEntry.Description,
					"client_id", clientId,
					"method", "sync",
				)
				err = usecase.timeEntryRepository.AddTimeEntry(timeEntry, tx)
				if err != nil {
					return err
				}

				changelogEntry := &model.ChangelogEntry{
					EntityType:      model.EntityTypeTimeEntry,
					EntityID:        timeEntry.ID,
					Operation:       model.OperationCreated,
					ChangedByUser:   userId,
					ChangedByClient: clientId,
				}
				err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			changelogEntry := &model.ChangelogEntry{
				EntityType:      model.EntityTypeTimeEntry,
				EntityID:        timeEntry.ID,
				Operation:       model.OperationUpdated,
				ChangedByUser:   userId,
				ChangedByClient: clientId,
			}
			err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (usecase *syncUsecase) processTimeEntryDeletions(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Processing time entry deletions", "count", len(timeEntries), "user_id", userId, "client_id", clientId, "method", "sync")
	for i := range timeEntries {
		timeEntry := &timeEntries[i]
		err := usecase.timeEntryRepository.DeleteTimeEntry(timeEntry, tx)
		if err != nil {
			if errors.Is(err, repository.ErrEntityNotFound) {
				slog.Warn("Time entry to be deleted not found, skipping",
					"time_entry_id", timeEntry.ID,
					"user_id", userId,
					"project_id", timeEntry.ProjectId,
					"description", timeEntry.Description,
					"client_id", clientId,
					"method", "sync",
				)
				continue
			} else {
				return err
			}
		}

		changelogEntry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeTimeEntry,
			EntityID:        timeEntry.ID,
			Operation:       model.OperationDeleted,
			ChangedByUser:   userId,
			ChangedByClient: clientId,
		}
		err = usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetChangedTimeEntries retrieves time entries that have changed since a specific changelog entry
func (usecase *syncUsecase) GetChangedTimeEntries(userId uuid.UUID, sinceTimeLogEntry int64, untilTimeLogEntry int64, excludeClientId string) (model.TimeEntrySyncResult, error) {
	return usecase.syncRepository.GetUpdatedTimeEntriesOfUser(userId, sinceTimeLogEntry, untilTimeLogEntry, excludeClientId)
}

// GetChangedProjects retrieves projects that have changed since a specific changelog entry
func (usecase *syncUsecase) GetChangedProjects(userId uuid.UUID, sinceTimeLogEntry int64, untilTimeLogEntry int64, excludeClientId string) (model.ProjectSyncResult, error) {
	return usecase.syncRepository.GetUpdatedProjectsOfUser(userId, sinceTimeLogEntry, untilTimeLogEntry, excludeClientId)
}

func (usecase *syncUsecase) GetProjectById(id uuid.UUID) (*model.Project, error) {
	return usecase.syncRepository.GetProjectById(id)
}

func (usecase *syncUsecase) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	return usecase.syncRepository.GetTimeEntryById(id)
}

func (usecase *syncUsecase) GetLatestChangelogEntryId() (int64, error) {
	return usecase.changelogRepository.GetLatestChangelogEntryId()
}

// processIssueDetection detects and resolves issues in time entry descriptions during sync
func (usecase *syncUsecase) processIssueDetection(ctx context.Context, timeEntry *model.TimeEntry, userId uuid.UUID) error {
	// Skip if no external integration usecase available (for backward compatibility)
	if usecase.externalUsecase == nil {
		return nil
	}

	// Use the centralized method from ExternalIntegrationUseCase
	return usecase.externalUsecase.ProcessTimeEntryIssueDetection(ctx, timeEntry, userId)
}

// mergeOpenTimeEntriesIfNeeded checks if there are existing open time entries for the same project
// and merges them if needed, using the earlier start time and the ID from the incoming entry
func (usecase *syncUsecase) mergeOpenTimeEntriesIfNeeded(newTimeEntry *model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	slog.Debug("Merging open time entries if needed", "time_entry_id", newTimeEntry.ID, "user_id", userId, "project_id", newTimeEntry.ProjectId, "client_id", clientId, "method", "sync")
	// Find existing open time entries for this project
	existingOpenEntries, getOpenErr := usecase.timeEntryRepository.GetOpenTimeEntriesForProject(userId, newTimeEntry.ProjectId, tx)
	if getOpenErr != nil {
		return getOpenErr
	}

	if len(existingOpenEntries) == 0 {
		// No existing open entries, just add the new one
		err := usecase.timeEntryRepository.AddTimeEntry(newTimeEntry, tx)
		if err != nil {
			return err
		}

		changelogEntry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeTimeEntry,
			EntityID:        newTimeEntry.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   userId,
			ChangedByClient: clientId,
		}
		return usecase.changelogRepository.AddChangelogEntry(changelogEntry, tx)
	}

	slog.Debug("Found existing open time entries to merge", "existing_count", len(existingOpenEntries), "time_entry_id", newTimeEntry.ID, "user_id", userId, "project_id", newTimeEntry.ProjectId, "client_id", clientId, "method", "sync")

	// Merge logic: use the most recent entry (latest start time) to avoid resurrecting old entries
	latestStartTime := newTimeEntry.StartTime
	var entryToKeep *model.TimeEntry = newTimeEntry
	var entriesToDelete []model.TimeEntry

	for _, existingEntry := range existingOpenEntries {
		if existingEntry.StartTime.Truncate(time.Second).After(latestStartTime.Truncate(time.Second)) {
			// Use the existing entry if it's more recent
			latestStartTime = existingEntry.StartTime
			entryToKeep = &existingEntry
			slog.Debug("Choosing existing entry to keep during merge", "kept_time_entry_id", entryToKeep.ID, "user_id", userId, "project_id", newTimeEntry.ProjectId, "client_id", clientId, "method", "sync")
			// Move the new entry to the delete list instead
			entriesToDelete = []model.TimeEntry{*newTimeEntry}
			// Add all other existing entries to delete list
			for _, otherEntry := range existingOpenEntries {
				if otherEntry.ID != existingEntry.ID {
					entriesToDelete = append(entriesToDelete, otherEntry)
				}
			}
			break
		} else {
			entriesToDelete = append(entriesToDelete, existingEntry)
		}
	}

	// Keep the start time of the most recent entry
	entryToKeep.StartTime = latestStartTime

	// If we're keeping an existing entry, respect the deleted flag from the incoming entry
	if entryToKeep != newTimeEntry {
		entryToKeep.Deleted = newTimeEntry.Deleted
	}

	// Combine descriptions if they exist and are different
	var descriptions []string

	// Always start with the description from the entry we're keeping
	if entryToKeep.Description != "" {
		descriptions = append(descriptions, entryToKeep.Description)
	}

	// Add descriptions from entries being deleted if they're different
	for _, entryToDelete := range entriesToDelete {
		if entryToDelete.Description != "" && entryToDelete.Description != entryToKeep.Description {
			// Avoid duplicates
			isDuplicate := false
			for _, existing := range descriptions {
				if existing == entryToDelete.Description {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				descriptions = append(descriptions, entryToDelete.Description)
			}
		}
	}

	// Update the description if we have multiple unique descriptions
	if len(descriptions) > 1 {
		entryToKeep.Description = strings.Join(descriptions, "; ")
	}

	// Delete the existing open entries
	for _, entryToDelete := range entriesToDelete {
		slog.Debug("Deleting duplicate open time entry during merge", "deleted_time_entry_id", entryToDelete.ID, "user_id", userId, "project_id", newTimeEntry.ProjectId, "client_id", clientId, "method", "sync")
		err := usecase.timeEntryRepository.DeleteTimeEntry(&entryToDelete, tx)
		if err != nil {
			return err
		}

		deleteChangelogEntry := &model.ChangelogEntry{
			EntityType:      model.EntityTypeTimeEntry,
			EntityID:        entryToDelete.ID,
			Operation:       model.OperationDeleted,
			ChangedByUser:   userId,
			ChangedByClient: "server-merge", // Use a special client ID so the mobile app receives it back
		}
		err = usecase.changelogRepository.AddChangelogEntry(deleteChangelogEntry, tx)
		if err != nil {
			return err
		}
	}

	// Check if we're keeping an existing entry or the new one
	isKeepingExisting := false
	for _, existingEntry := range existingOpenEntries {
		if existingEntry.ID == entryToKeep.ID {
			isKeepingExisting = true
			break
		}
	}

	var changelogOperation model.Operation
	var err error
	if isKeepingExisting {
		slog.Debug("Updating existing time entry during merge", "kept_time_entry_id", entryToKeep.ID, "user_id", userId, "project_id", newTimeEntry.ProjectId, "client_id", clientId, "method", "sync")
		// Update the existing entry with merged data
		err = usecase.timeEntryRepository.UpdateTimeEntry(entryToKeep, tx)
		changelogOperation = model.OperationUpdated
	} else {
		slog.Debug("Adding new time entry during merge", "kept_time_entry_id", entryToKeep.ID, "user_id", userId, "project_id", newTimeEntry.ProjectId, "client_id", clientId, "method", "sync")
		// Add the new entry
		err = usecase.timeEntryRepository.AddTimeEntry(entryToKeep, tx)
		changelogOperation = model.OperationCreated
	}

	if err != nil {
		return err
	}

	createChangelogEntry := &model.ChangelogEntry{
		EntityType:      model.EntityTypeTimeEntry,
		EntityID:        entryToKeep.ID,
		Operation:       changelogOperation,
		ChangedByUser:   userId,
		ChangedByClient: "server-merge", // Use a special client ID so the mobile app receives it back
	}
	return usecase.changelogRepository.AddChangelogEntry(createChangelogEntry, tx)
}
