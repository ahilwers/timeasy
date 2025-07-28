package usecase

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type SyncUsecase interface {
	UpdateAndDeleteData(data model.SyncData, userId uuid.UUID, clientId string) error
	GetChangedTimeEntries(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.TimeEntrySyncResult, error)
	GetChangedProjects(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.ProjectSyncResult, error)
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
}

type syncUsecase struct {
	syncRepository     repository.SyncRepository
	changelogRepository repository.ChangelogRepository
	projectRepository  repository.ProjectRepository
	timeEntryRepository repository.TimeEntryRepository
}

func NewSyncUsecase(
	syncRepository repository.SyncRepository,
	changelogRepository repository.ChangelogRepository,
	projectRepository repository.ProjectRepository,
	timeEntryRepository repository.TimeEntryRepository,
) SyncUsecase {
	return &syncUsecase{
		syncRepository:     syncRepository,
		changelogRepository: changelogRepository,
		projectRepository:  projectRepository,
		timeEntryRepository: timeEntryRepository,
	}
}

// UpdateAndDeleteData processes the sync data by creating, updating, and deleting entries
// and adds appropriate changelog entries for each operation
func (usecase *syncUsecase) UpdateAndDeleteData(data model.SyncData, userId uuid.UUID, clientId string) error {
	// Begin a transaction for project operations
	projectTx, err := usecase.projectRepository.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			projectTx.Rollback()
		}
	}()

	// Begin a transaction for time entry operations
	timeEntryTx, err := usecase.timeEntryRepository.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			timeEntryTx.Rollback()
		}
	}()

	// Process projects
	err = usecase.processProjects(data, userId, clientId, projectTx)
	if err != nil {
		return err
	}

	// Process time entries
	err = usecase.processTimeEntries(data, userId, clientId, timeEntryTx)
	if err != nil {
		return err
	}

	// Commit the transactions if everything was successful
	err = projectTx.Commit()
	if err != nil {
		return err
	}

	err = timeEntryTx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// processProjects handles the creation, update, and deletion of projects
func (usecase *syncUsecase) processProjects(data model.SyncData, userId uuid.UUID, clientId string, tx model.Transaction) error {
	// Process projects to be created
	err := usecase.processProjectCreations(data.ProjectsToBeCreated, userId, clientId, tx)
	if err != nil {
		return err
	}

	// Process projects to be updated
	err = usecase.processProjectUpdates(data.ProjectsToBeUpdated, userId, clientId, tx)
	if err != nil {
		return err
	}

	// Process projects to be deleted
	err = usecase.processProjectDeletions(data.ProjectsToBeDeleted, userId, clientId, tx)
	if err != nil {
		return err
	}

	return nil
}

// processProjectCreations handles the creation of projects and adds changelog entries
func (usecase *syncUsecase) processProjectCreations(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.AddProject(project, tx)
		if err != nil {
			return err
		}

		// Add changelog entry for project creation
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

// processProjectUpdates handles the updating of projects and adds changelog entries
func (usecase *syncUsecase) processProjectUpdates(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.UpdateProject(project, tx)
		if err != nil {
			return err
		}

		// Add changelog entry for project update
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

// processProjectDeletions handles the deletion of projects and adds changelog entries
func (usecase *syncUsecase) processProjectDeletions(projects []model.Project, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range projects {
		project := &projects[i]
		err := usecase.projectRepository.DeleteProject(project, tx)
		if err != nil {
			return err
		}

		// Add changelog entry for project deletion
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

// processTimeEntries handles the creation, update, and deletion of time entries
func (usecase *syncUsecase) processTimeEntries(data model.SyncData, userId uuid.UUID, clientId string, tx model.Transaction) error {
	// Process time entries to be created
	err := usecase.processTimeEntryCreations(data.TimeEntriesToBeCreated, userId, clientId, tx)
	if err != nil {
		return err
	}

	// Process time entries to be updated
	err = usecase.processTimeEntryUpdates(data.TimeEntriesToBeUpdated, userId, clientId, tx)
	if err != nil {
		return err
	}

	// Process time entries to be deleted
	err = usecase.processTimeEntryDeletions(data.TimeEntriesToBeDeleted, userId, clientId, tx)
	if err != nil {
		return err
	}

	return nil
}

// processTimeEntryCreations handles the creation of time entries and adds changelog entries
func (usecase *syncUsecase) processTimeEntryCreations(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range timeEntries {
		timeEntry := &timeEntries[i]
		err := usecase.timeEntryRepository.AddTimeEntry(timeEntry, tx)
		if err != nil {
			return err
		}

		// Add changelog entry for time entry creation
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
	return nil
}

// processTimeEntryUpdates handles the updating of time entries and adds changelog entries
func (usecase *syncUsecase) processTimeEntryUpdates(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range timeEntries {
		timeEntry := &timeEntries[i]
		err := usecase.timeEntryRepository.UpdateTimeEntry(timeEntry, tx)
		if err != nil {
			return err
		}

		// Add changelog entry for time entry update
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

// processTimeEntryDeletions handles the deletion of time entries and adds changelog entries
func (usecase *syncUsecase) processTimeEntryDeletions(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range timeEntries {
		timeEntry := &timeEntries[i]
		err := usecase.timeEntryRepository.DeleteTimeEntry(timeEntry, tx)
		if err != nil {
			return err
		}

		// Add changelog entry for time entry deletion
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

// GetChangedTimeEntries retrieves time entries that have changed since a specific time
func (usecase *syncUsecase) GetChangedTimeEntries(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.TimeEntrySyncResult, error) {
	return usecase.syncRepository.GetUpdatedTimeEntriesOfUser(userId, sinceTimeLogEntry, excludeClientId)
}

// GetChangedProjects retrieves projects that have changed since a specific time
func (usecase *syncUsecase) GetChangedProjects(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.ProjectSyncResult, error) {
	return usecase.syncRepository.GetUpdatedProjectsOfUser(userId, sinceTimeLogEntry, excludeClientId)
}

// GetProjectById retrieves a project by its ID
func (usecase *syncUsecase) GetProjectById(id uuid.UUID) (*model.Project, error) {
	return usecase.syncRepository.GetProjectById(id)
}

// GetTimeEntryById retrieves a time entry by its ID
func (usecase *syncUsecase) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	return usecase.syncRepository.GetTimeEntryById(id)
}
