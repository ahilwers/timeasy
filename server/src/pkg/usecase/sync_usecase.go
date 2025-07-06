package usecase

import (
	"github.com/gofrs/uuid"
	"time"
	"timeasy-server/pkg/domain/model"
)

type SyncUsecase interface {
	UpdateAndDeleteData(data model.SyncData) error
	GetChangedTimeEntries(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error)
	GetChangedProjects(userId uuid.UUID, sinceWhen time.Time) ([]model.Project, error)
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
}

type syncUsecase struct {
}

func NewSyncUsecase() SyncUsecase {
	return &syncUsecase{}
}

func (usecase *syncUsecase) UpdateAndDeleteData(data model.SyncData) error {
	return nil
}

func (tu *syncUsecase) GetChangedTimeEntries(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error) {
	return nil, nil
}

func (tu *syncUsecase) GetChangedProjects(userId uuid.UUID, sinceWhen time.Time) ([]model.Project, error) {
	return nil, nil
}

func (usecase *syncUsecase) GetProjectById(id uuid.UUID) (*model.Project, error) {
	return nil, nil
}

func (usecase *syncUsecase) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	return nil, nil
}
