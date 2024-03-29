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
	AddTimeEntryList(timeEntryList []model.TimeEntry) error
	UpdateTimeEntry(timeEntry *model.TimeEntry) error
	UpdateTimeEntryList(timeEntry []model.TimeEntry) error
	DeleteTimeEntry(id uuid.UUID) error
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

func (tu *timeEntryUsecase) AddTimeEntryList(timeEntryList []model.TimeEntry) error {
	for _, timeEntry := range timeEntryList {
		err := tu.checkEntry(&timeEntry)
		if err != nil {
			return err
		}
	}
	return tu.repo.AddTimeEntryList(timeEntryList)
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

func (tu *timeEntryUsecase) UpdateTimeEntryList(timeEntryList []model.TimeEntry) error {
	for _, timeEntry := range timeEntryList {
		err := tu.checkEntry(&timeEntry)
		if err != nil {
			return err
		}
	}
	return tu.repo.UpdateTimeEntryList(timeEntryList)
}

func (tu *timeEntryUsecase) DeleteTimeEntry(id uuid.UUID) error {
	timeEntry, err := tu.GetTimeEntryById(id)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("timeEntry with id %v does not exist", id))
	}
	return tu.repo.DeleteTimeEntry(timeEntry)
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
