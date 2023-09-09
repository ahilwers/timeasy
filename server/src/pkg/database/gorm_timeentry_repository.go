package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type gormTimeEntryRepository struct {
	db *gorm.DB
}

func NewGormTimeEntryRepository(database *gorm.DB) repository.TimeEntryRepository {
	return &gormTimeEntryRepository{
		db: database,
	}
}

func (repo *gormTimeEntryRepository) AddTimeEntry(timeEntry *model.TimeEntry) error {
	if err := repo.db.Create(timeEntry).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTimeEntryRepository) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	var timeEntry model.TimeEntry
	if err := repo.db.First(&timeEntry, id).Error; err != nil {
		return nil, err
	}
	return &timeEntry, nil
}

func (repo *gormTimeEntryRepository) UpdateTimeEntry(timeEntry *model.TimeEntry) error {
	if err := repo.db.Save(timeEntry).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTimeEntryRepository) DeleteTimeEntry(timeEntry *model.TimeEntry) error {
	if err := repo.db.Model(&timeEntry).Update("deleted", true).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTimeEntryRepository) GetAllTimeEntries() ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Order("end_time desc").Find(&timeEntries, "deleted=?", false).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}

func (repo *gormTimeEntryRepository) GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Order("end_time desc").Find(&timeEntries, "deleted=? AND user_id=?",
		false, userId).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}

func (repo *gormTimeEntryRepository) GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Order("end_time desc").Find(&timeEntries, " deleted=? AND user_id=? AND project_id=?",
		false, userId, projectId).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}
