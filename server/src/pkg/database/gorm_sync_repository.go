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
		err := repo.SaveProject(tx, &project)
		if err != nil {
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

func (repo *gormSyncRepository) SaveProject(tx *gorm.DB, project *model.Project) error {
	existingProject, _ := repo.GetProjectById(project.ID)
	// Don't overwrite created date and ownership of existing projects
	if existingProject != nil {
		project.UserId = existingProject.UserId
		project.CreatedAt = existingProject.CreatedAt
	}
	if err := tx.Save(&project).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormSyncRepository) GetProjectById(id uuid.UUID) (*model.Project, error) {
	var project model.Project
	if err := repo.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (repo *gormSyncRepository) updateAndDeleteTimeEntries(tx *gorm.DB, data model.SyncData) error {
	for _, timeEntry := range data.TimeEntriesToBeUpdated {
		err := repo.SaveTimeEntry(tx, &timeEntry)
		if err != nil {
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

func (repo *gormSyncRepository) SaveTimeEntry(tx *gorm.DB, project *model.TimeEntry) error {
	existingTimeEntry, _ := repo.GetTimeEntryById(project.ID)
	// Don't overwrite created date and ownership of existing time entries
	if existingTimeEntry != nil {
		project.UserId = existingTimeEntry.UserId
		project.CreatedAt = existingTimeEntry.CreatedAt
	}
	if err := tx.Save(&project).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormSyncRepository) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	var timeEntry model.TimeEntry
	if err := repo.db.First(&timeEntry, id).Error; err != nil {
		return nil, err
	}
	return &timeEntry, nil
}

func (repo *gormSyncRepository) GetUpdatedTimeEntriesOfUser(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error) {
	var updatedEntries []model.TimeEntry
	if err := repo.db.Unscoped().Order("start_time desc").Order("end_time desc").Find(&updatedEntries, "user_id=? AND (updated_at >= ? OR created_at >= ? OR deleted_at >= ?)", userId, sinceWhen, sinceWhen, sinceWhen).Error; err != nil {
		return nil, err
	}
	return updatedEntries, nil
}

func (repo *gormSyncRepository) GetUpdatedProjectsOfUser(userId uuid.UUID, sinceWhen time.Time) ([]model.Project, error) {
	var updatedProjects []model.Project
	if err := repo.db.Unscoped().Order("name").Find(&updatedProjects, "user_id=? AND (updated_at >= ? OR created_at >= ? OR deleted_at >= ?)", userId, sinceWhen, sinceWhen, sinceWhen).Error; err != nil {
		return nil, err
	}
	return updatedProjects, nil
}
