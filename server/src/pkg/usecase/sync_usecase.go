package usecase

import (
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
}

func NewSyncUsecase(
	syncRepository repository.SyncRepository,
	changelogRepository repository.ChangelogRepository,
	projectRepository repository.ProjectRepository,
	timeEntryRepository repository.TimeEntryRepository,
) SyncUsecase {
	return &syncUsecase{
		syncRepository:      syncRepository,
		changelogRepository: changelogRepository,
		projectRepository:   projectRepository,
		timeEntryRepository: timeEntryRepository,
	}
}

// UpdateAndDeleteData processes the sync data by creating, updating, and deleting entries
// and adds appropriate changelog entries for each operation
func (usecase *syncUsecase) UpdateAndDeleteData(data model.SyncData, userId uuid.UUID, clientId string) error {
	projectTx, err := usecase.projectRepository.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			projectTx.Rollback()
		}
	}()

	timeEntryTx, err := usecase.timeEntryRepository.BeginTransaction()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			timeEntryTx.Rollback()
		}
	}()

	err = usecase.processProjects(data, userId, clientId, projectTx)
	if err != nil {
		return err
	}

	err = usecase.processTimeEntries(data, userId, clientId, timeEntryTx)
	if err != nil {
		return err
	}

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
	return nil
}

func (usecase *syncUsecase) processTimeEntryUpdates(timeEntries []model.TimeEntry, userId uuid.UUID, clientId string, tx model.Transaction) error {
	for i := range timeEntries {
		timeEntry := &timeEntries[i]
		err := usecase.timeEntryRepository.UpdateTimeEntry(timeEntry, tx)
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
