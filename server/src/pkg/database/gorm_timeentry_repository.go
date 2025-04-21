package database

import (
	"time"
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

func (repo *gormTimeEntryRepository) AddTimeEntryList(timeEntryList []model.TimeEntry) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for _, timeEntry := range timeEntryList {
			if err := tx.Create(&timeEntry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *gormTimeEntryRepository) GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error) {
	var timeEntry model.TimeEntry
	if err := repo.db.First(&timeEntry, id).Error; err != nil {
		return nil, err
	}
	return &timeEntry, nil
}

func (repo *gormTimeEntryRepository) GetLastOpenTimeEntry(userId uuid.UUID) (*model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Find(&timeEntries, "user_id=? AND end_time=?", userId, "0001-01-01 00:00:00").Error; err != nil {
		return nil, err
	}
	if len(timeEntries) == 0 {
		return nil, nil
	}
	return &timeEntries[0], nil
}

func (repo *gormTimeEntryRepository) UpdateTimeEntry(timeEntry *model.TimeEntry) error {
	if err := repo.db.Save(timeEntry).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTimeEntryRepository) UpdateTimeEntryList(timeEntryList []model.TimeEntry) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		for _, timeEntry := range timeEntryList {
			if err := tx.Save(&timeEntry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (repo *gormTimeEntryRepository) DeleteTimeEntry(timeEntry *model.TimeEntry) error {
	if err := repo.db.Delete(timeEntry).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTimeEntryRepository) GetAllTimeEntries() ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Order("end_time desc").Find(&timeEntries).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}

func (repo *gormTimeEntryRepository) GetAllTimeEntriesOfUser(userId uuid.UUID) ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Order("end_time desc").Find(&timeEntries, "user_id=?", userId).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}

func (repo *gormTimeEntryRepository) GetAllTimeEntriesOfUserAndProject(userId uuid.UUID, projectId uuid.UUID) ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	if err := repo.db.Order("start_time desc").Order("end_time desc").Find(&timeEntries, "user_id=? AND project_id=?",
		userId, projectId).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}

func (repo *gormTimeEntryRepository) GetTimeEntriesOfUserAndProjectBetweenDates(userId uuid.UUID, projectId uuid.UUID, startDate time.Time, endDate time.Time) ([]model.TimeEntry, error) {
	var timeEntries []model.TimeEntry
	query := repo.db.Where("user_id=?", userId)
	if projectId != uuid.Nil {
		query = query.Where("project_id=?", projectId)
	}
	query = query.Where("DATE(start_time)>=? AND DATE(end_time)<=?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err := query.Order("start_time desc").Order("end_time desc").Find(&timeEntries).Error; err != nil {
		return nil, err
	}
	return timeEntries, nil
}
