package usecase

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
)

type SyncUsecase interface {
	UpdateAndDeleteData(data model.SyncData) error
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
