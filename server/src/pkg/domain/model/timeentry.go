package model

import (
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type TimeEntry struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;"`
	UserId      uuid.UUID `gorm:"type:uuid;"`
	ProjectId   uuid.UUID `gorm:"type:uuid;"`
	Project     Project
	StartTime   time.Time `gorm:"type:timestamp;"` // db: timestamp without time zone
	EndTime     time.Time `gorm:"type:timestamp;"` // db: timestamp without time zone
	Description string
}

func (timeEntry *TimeEntry) BeforeCreate(db *gorm.DB) error {
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
