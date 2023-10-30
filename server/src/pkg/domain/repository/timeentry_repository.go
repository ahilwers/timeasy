package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type TimeEntryRepository interface {
	AddTimeEntry(project *model.TimeEntry) error
	AddTimeEntryList(timeEntryList []model.TimeEntry) error
	UpdateTimeEntry(timeEntry *model.TimeEntry) error
	UpdateTimeEntryList(timeEntryList []model.TimeEntry) error
	DeleteTimeEntry(project *model.TimeEntry) error
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
	GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error)
	GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error)
}
