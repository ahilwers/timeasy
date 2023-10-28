package database

import (
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
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

func (repo *gormSyncRepository) GetUpdatedTimeEntriesOfUser(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error) {
	var updatedEntries []model.TimeEntry
	if err := repo.db.Unscoped().Order("start_time desc").Order("end_time desc").Find(&updatedEntries, "user_id=? AND (updated_at >= ? OR created_at >= ? OR deleted_at >= ?)", userId, sinceWhen, sinceWhen, sinceWhen).Error; err != nil {
		return nil, err
	}
	return updatedEntries, nil
}
