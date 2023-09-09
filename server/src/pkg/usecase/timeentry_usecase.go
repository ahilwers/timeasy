package usecase

import (
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
	AddTimeEntry(timeEntry *model.TimeEntry) error
	UpdateTimeEntry(timeEntry *model.TimeEntry) error
	DeleteTimeEntry(id uuid.UUID) error
	GetChangedEntries(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error)
}

type timeEntryUsecase struct {
	repo           repository.TimeEntryRepository
	projectUsecase ProjectUsecase
}

func NewTimeEntryUsecase(repo repository.TimeEntryRepository, projectUsecase ProjectUsecase) TimeEntryUsecase {
	return &timeEntryUsecase{
		repo:           repo,
		projectUsecase: projectUsecase,
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

func (tu *timeEntryUsecase) AddTimeEntry(timeEntry *model.TimeEntry) error {
	err := tu.checkEntry(timeEntry)
	if err != nil {
		return err
	}
	return tu.repo.AddTimeEntry(timeEntry)
}

func (tu *timeEntryUsecase) UpdateTimeEntry(timeEntry *model.TimeEntry) error {
	_, err := tu.GetTimeEntryById(timeEntry.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", timeEntry.ID))
	}
	err = tu.checkEntry(timeEntry)
	if err != nil {
		return err
	}
	return tu.repo.UpdateTimeEntry(timeEntry)
}

func (tu *timeEntryUsecase) DeleteTimeEntry(id uuid.UUID) error {
	timeEntry, err := tu.GetTimeEntryById(id)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", id))
	}
	return tu.repo.DeleteTimeEntry(timeEntry)
}

func (tu *timeEntryUsecase) GetChangedEntries(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error) {
	return tu.repo.GetUpdatedTimeEntriesOfUser(userId, sinceWhen)
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
	return nil
}

func (tu *timeEntryUsecase) checkUser(timeEntry *model.TimeEntry) error {
	if timeEntry.UserId == uuid.Nil {
		return NewEntityIncompleteError("the user id must not be empty")
	}
	return nil
}

func (tu *timeEntryUsecase) checkProject(timeEntry *model.TimeEntry) error {
	if timeEntry.ProjectId == uuid.Nil {
		return NewEntityIncompleteError("the project id must not be empty")
	}
	_, err := tu.projectUsecase.GetProjectById(timeEntry.ProjectId)
	if err != nil {
		return NewProjectNotFoundError(timeEntry.ProjectId)
	}
	return nil
}
