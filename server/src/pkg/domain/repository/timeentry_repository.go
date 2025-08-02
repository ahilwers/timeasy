package repository

import (
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type TimeEntryRepository interface {
	model.TransactionHandler
	AddTimeEntry(project *model.TimeEntry, tx model.Transaction) error
	UpdateTimeEntry(timeEntry *model.TimeEntry, tx model.Transaction) error
	DeleteTimeEntry(project *model.TimeEntry, tx model.Transaction) error
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
	GetLastOpenTimeEntry(userId uuid.UUID) (*model.TimeEntry, error)
	GetAllTimeEntries() ([]model.TimeEntry, error)
	GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error)
	GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error)
	GetTimeEntriesOfUserAndProjectBetweenDates(userId uuid.UUID, projectId uuid.UUID, startDate time.Time, endDate time.Time) ([]model.TimeEntry, error)
}
