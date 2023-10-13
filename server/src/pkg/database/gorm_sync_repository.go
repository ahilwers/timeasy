package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"gorm.io/gorm"
)

type gormSyncRepository struct {
	db *gorm.DB
}

func NewGormSyncRepository(database *gorm.DB) repository.SyncRepository {
	return &gormSyncRepository{
		db: database,
	}
}

func (repo *gormSyncRepository) UpdateAndDeleteData(data model.SyncData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		err := repo.updateAndDeleteProjects(tx, data)
		if err != nil {
			return err
		}
		err = repo.updateAndDeleteTimeEntries(tx, data)
		if err != nil {
			return err
		}
		return nil
	})
}

func (repo *gormSyncRepository) updateAndDeleteProjects(tx *gorm.DB, data model.SyncData) error {
	for _, project := range data.ProjectsToBeUpdated {
		if err := tx.Save(&project).Error; err != nil {
			return err
		}
	}
	for _, project := range data.ProjectsToBeDeleted {
		if err := tx.Delete(&project).Error; err != nil {
			return err
		}
	}
	return nil
}

func (repo *gormSyncRepository) updateAndDeleteTimeEntries(tx *gorm.DB, data model.SyncData) error {
	for _, timeEntry := range data.TimeEntriesToBeUpdated {
		if err := tx.Save(&timeEntry).Error; err != nil {
			return err
		}
	}
	for _, timeEntry := range data.TimeEntriesToBeDeleted {
		if err := tx.Delete(&timeEntry).Error; err != nil {
			return err
		}
	}
	return nil
}
