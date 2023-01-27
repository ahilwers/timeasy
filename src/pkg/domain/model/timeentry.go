package model

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type TimeEntry struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;"`
	UserId      uuid.UUID
	ProjectId   uuid.UUID
	StartTime   time.Time
	EndTime     time.Time
	Description string
}

func (timeEntry *TimeEntry) BeforeCreate(db *gorm.DB) error {
	time.Now()
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	timeEntry.ID = id
	if timeEntry.StartTime.IsZero() {
		timeEntry.StartTime = time.Now().UTC()
	}
	return nil
}
