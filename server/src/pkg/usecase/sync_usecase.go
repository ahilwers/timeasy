package usecase

import (
	"context"
	"log/slog"
	"strings"
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
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.AddProject(project, tx)
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
	}
	return nil
}

func (usecase *syncUsecase) processProjectUpdates(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.UpdateProject(project, tx)
		if err != nil {
			return err
		}

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
	return nil
}

func (usecase *syncUsecase) processProjectDeletions(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.DeleteProject(project, tx)
		if err != nil {
			return err
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
	return nil
}

func (usecase *syncUsecase) processTimeEntryUpdates(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
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
	}
	return nil
}

func (usecase *syncUsecase) processTimeEntryDeletions(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range timeEntries {
		timeEntry := &timeEntries[i]
		err := usecase.timeEntryRepository.DeleteTimeEntry(timeEntry, tx)
		if err != nil {
			return err
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
	// Find existing open time entries for this project
	existingOpenEntries, err := usecase.timeEntryRepository.GetOpenTimeEntriesForProject(userId, newTimeEntry.ProjectId, tx)
	if err != nil {
		return err
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

	// Merge logic: use the earliest start time and the ID from the new entry
	earliestStartTime := newTimeEntry.StartTime
	var entryToKeep *model.TimeEntry = newTimeEntry
	var entriesToDelete []model.TimeEntry

	for _, existingEntry := range existingOpenEntries {
		if existingEntry.StartTime.Before(earliestStartTime) {
			earliestStartTime = existingEntry.StartTime
		}
		entriesToDelete = append(entriesToDelete, existingEntry)
	}

	// Update the new entry with the earliest start time
	entryToKeep.StartTime = earliestStartTime
	
	// Combine descriptions if they exist and are different
	var descriptions []string
	if newTimeEntry.Description != "" {
		descriptions = append(descriptions, newTimeEntry.Description)
	}
	for _, existingEntry := range existingOpenEntries {
		if existingEntry.Description != "" && existingEntry.Description != newTimeEntry.Description {
			descriptions = append(descriptions, existingEntry.Description)
		}
	}
	if len(descriptions) > 1 {
		entryToKeep.Description = strings.Join(descriptions, "; ")
	} else if len(descriptions) == 1 {
		entryToKeep.Description = descriptions[0]
	}

	// Delete the existing open entries
	for _, entryToDelete := range entriesToDelete {
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

	// Add the merged entry
	err = usecase.timeEntryRepository.AddTimeEntry(entryToKeep, tx)
	if err != nil {
		return err
	}

	createChangelogEntry := &model.ChangelogEntry{
		EntityType:      model.EntityTypeTimeEntry,
		EntityID:        entryToKeep.ID,
		Operation:       model.OperationCreated,
		ChangedByUser:   userId,
		ChangedByClient: "server-merge", // Use a special client ID so the mobile app receives it back
	}
	return usecase.changelogRepository.AddChangelogEntry(createChangelogEntry, tx)
}
