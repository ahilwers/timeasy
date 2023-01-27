package usecase

import (
	"fmt"
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
}

type timeEntryUsecase struct {
	repo repository.TimeEntryRepository
}

func NewTimeEntryUsecase(repo repository.TimeEntryRepository) TimeEntryUsecase {
	return &timeEntryUsecase{
		repo: repo,
	}
}

func (pu *timeEntryUsecase) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	return pu.repo.GetTimeEntryById(id)
}

func (pu *timeEntryUsecase) GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error) {
	return pu.repo.GetAllTimeEntriesOfUser(userId)
}

func (pu *timeEntryUsecase) GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error) {
	return pu.repo.GetAllTimeEntriesOfUserAndProject(userId, projectId)
}

func (pu *timeEntryUsecase) AddTimeEntry(timeEntry *model.TimeEntry) error {
	if timeEntry.UserId == uuid.Nil {
		return NewEntityIncompleteError("the user id must not be empty")
	}
	return pu.repo.AddTimeEntry(timeEntry)
}

func (pu *timeEntryUsecase) UpdateTimeEntry(timeEntry *model.TimeEntry) error {
	if timeEntry.UserId == uuid.Nil {
		return NewEntityIncompleteError("the user id must not be empty")
	}
	_, err := pu.GetTimeEntryById(timeEntry.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", timeEntry.ID))
	}
	return pu.repo.UpdateTimeEntry(timeEntry)
}

func (pu *timeEntryUsecase) DeleteTimeEntry(id uuid.UUID) error {
	timeEntry, err := pu.GetTimeEntryById(id)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", id))
	}
	return pu.repo.DeleteTimeEntry(timeEntry)
}
