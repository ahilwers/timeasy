package usecase

import (
	"errors"
	"fmt"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type TimeEntryUsecase interface {
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
	GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error)
	GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error)
	GetTimeEntriesOfUserAndProjectBetweenDates(userId uuid.UUID, projectId uuid.UUID, startDate time.Time, endDate time.Time) ([]model.TimeEntry, error)
	GetLastOpenTimeEntry(userId uuid.UUID) (*model.TimeEntry, error)
	AddTimeEntry(timeEntry *model.TimeEntry, userId uuid.UUID, clientId string) error
	AddTimeEntryList(timeEntryList []model.TimeEntry, userId uuid.UUID, clientId string) error
	UpdateTimeEntry(timeEntry *model.TimeEntry, userId uuid.UUID, clientId string) error
	UpdateTimeEntryList(timeEntry []model.TimeEntry, userId uuid.UUID, clientId string) error
	DeleteTimeEntry(id uuid.UUID, userId uuid.UUID, clientId string) error
	GetLastActivityTimeForProject(projectId uuid.UUID) (time.Time, error)
	GetProjectsWithRecentActivity(since time.Time) ([]uuid.UUID, error)
}

type timeEntryUsecase struct {
	repo           repository.TimeEntryRepository
	projectUsecase ProjectUsecase
	changelogRepo  repository.ChangelogRepository
}

func NewTimeEntryUsecase(repo repository.TimeEntryRepository, projectUsecase ProjectUsecase, changelogRepo repository.ChangelogRepository) TimeEntryUsecase {
	return &timeEntryUsecase{
		repo:           repo,
		projectUsecase: projectUsecase,
		changelogRepo:  changelogRepo,
	}
}

func (tu *timeEntryUsecase) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	entry, err := tu.repo.GetTimeEntryById(id)
	if err != nil {
		return nil, NewEntityNotFoundError(fmt.Sprintf("timeentry with if %v does not exist", id))
	}
	return entry, nil
}

func (tu *timeEntryUsecase) GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error) {
	return tu.repo.GetAllTimeEntriesOfUser(userId)
}

func (tu *timeEntryUsecase) GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error) {
	return tu.repo.GetAllTimeEntriesOfUserAndProject(userId, projectId)
}

func (tu *timeEntryUsecase) GetTimeEntriesOfUserAndProjectBetweenDates(userId uuid.UUID, projectId uuid.UUID, startDate time.Time, endDate time.Time) ([]model.TimeEntry, error) {
	return tu.repo.GetTimeEntriesOfUserAndProjectBetweenDates(userId, projectId, startDate, endDate)
}

func (tu *timeEntryUsecase) GetLastOpenTimeEntry(userId uuid.UUID) (*model.TimeEntry, error) {
	timeEntry, err := tu.repo.GetLastOpenTimeEntry(userId)
	if err != nil {
		if errors.Is(err, repository.ErrEntityNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return timeEntry, nil
}

func (tu *timeEntryUsecase) AddTimeEntry(timeEntry *model.TimeEntry, userId uuid.UUID, clientId string) error {
	err := tu.checkEntry(timeEntry)
	if err != nil {
		return err
	}
	tx, err := tu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	err = tu.repo.AddTimeEntry(timeEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeTimeEntry,
		EntityID:        timeEntry.ID,
		Operation:       model.OperationCreated,
		ChangedByUser:   userId,
		ChangedByClient: clientId,
	}
	err = tu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (tu *timeEntryUsecase) AddTimeEntryList(timeEntryList []model.TimeEntry, userId uuid.UUID, clientId string) error {
	for _, timeEntry := range timeEntryList {
		err := tu.checkEntry(&timeEntry)
		if err != nil {
			return err
		}
	}
	tx, err := tu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	for _, timeEntry := range timeEntryList {
		err := tu.repo.AddTimeEntry(&timeEntry, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		changelogEntry := model.ChangelogEntry{
			EntityType:      model.EntityTypeTimeEntry,
			EntityID:        timeEntry.ID,
			Operation:       model.OperationCreated,
			ChangedByUser:   userId,
			ChangedByClient: clientId,
		}
		err = tu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (tu *timeEntryUsecase) UpdateTimeEntry(timeEntry *model.TimeEntry, userId uuid.UUID, clientId string) error {
	_, err := tu.GetTimeEntryById(timeEntry.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", timeEntry.ID))
	}
	err = tu.checkEntry(timeEntry)
	if err != nil {
		return err
	}
	tx, err := tu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	err = tu.repo.UpdateTimeEntry(timeEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeTimeEntry,
		EntityID:        timeEntry.ID,
		Operation:       model.OperationUpdated,
		ChangedByUser:   userId,
		ChangedByClient: clientId,
	}
	err = tu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (tu *timeEntryUsecase) UpdateTimeEntryList(timeEntry []model.TimeEntry, userId uuid.UUID, clientId string) error {
	for _, timeEntry := range timeEntry {
		err := tu.checkEntry(&timeEntry)
		if err != nil {
			return err
		}
	}
	tx, err := tu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	for _, timeEntry := range timeEntry {
		err = tu.repo.UpdateTimeEntry(&timeEntry, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		changelogEntry := model.ChangelogEntry{
			EntityType:      model.EntityTypeTimeEntry,
			EntityID:        timeEntry.ID,
			Operation:       model.OperationUpdated,
			ChangedByUser:   userId,
			ChangedByClient: clientId,
		}
		err = tu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (tu *timeEntryUsecase) DeleteTimeEntry(id uuid.UUID, userId uuid.UUID, clientId string) error {
	timeEntry, err := tu.GetTimeEntryById(id)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", id))
	}
	tx, err := tu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	err = tu.repo.DeleteTimeEntry(timeEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeTimeEntry,
		EntityID:        timeEntry.ID,
		Operation:       model.OperationDeleted,
		ChangedByUser:   userId,
		ChangedByClient: clientId,
	}
	err = tu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (tu *timeEntryUsecase) checkEntry(timeEntry *model.TimeEntry) error {
	err := tu.checkUser(timeEntry)
	if err != nil {
		return err
	}
	err = tu.checkProject(timeEntry)
	if err != nil {
		return err
	}
	if timeEntry.StartTime.IsZero() {
		timeEntry.StartTime = time.Now().UTC()
	}
	return nil
}

func (tu *timeEntryUsecase) checkUser(timeEntry *model.TimeEntry) error {
	if timeEntry.UserId == uuid.Nil {
		return NewEntityIncompleteError(fmt.Sprintf("the user id of time entry %v must not be empty", timeEntry.ID))
	}
	return nil
}

func (tu *timeEntryUsecase) checkProject(timeEntry *model.TimeEntry) error {
	if timeEntry.ProjectId == uuid.Nil {
		return NewEntityIncompleteError(fmt.Sprintf("the project id of time entry %v must not be empty", timeEntry.ID))
	}
	_, err := tu.projectUsecase.GetProjectById(timeEntry.ProjectId)
	if err != nil {
		return NewProjectNotFoundError(timeEntry.ProjectId)
	}
	return nil
}

func (tu *timeEntryUsecase) GetLastActivityTimeForProject(projectId uuid.UUID) (time.Time, error) {
	return tu.repo.GetLastActivityTimeForProject(projectId)
}

func (tu *timeEntryUsecase) GetProjectsWithRecentActivity(since time.Time) ([]uuid.UUID, error) {
	return tu.repo.GetProjectsWithRecentActivity(since)
}
