package usecase

import (
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type SyncUsecase interface {
	UpdateAndDeleteData(data model.SyncData) error
	GetChangedTimeEntries(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error)
}

type syncUsecase struct {
	repo repository.SyncRepository
}

func NewSyncUsecase(repo repository.SyncRepository) SyncUsecase {
	return &syncUsecase{
		repo: repo,
	}
}

func (usecase *syncUsecase) UpdateAndDeleteData(data model.SyncData) error {
	return usecase.repo.UpdateAndDeleteData(data)
}

func (tu *syncUsecase) GetChangedTimeEntries(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error) {
	return tu.repo.GetUpdatedTimeEntriesOfUser(userId, sinceWhen)
}
