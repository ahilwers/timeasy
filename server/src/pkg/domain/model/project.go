package model

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	ID     uuid.UUID  `gorm:"type:uuid;primaryKey;" json:"id"`
	Name   string     `json:"name"`
	UserId uuid.UUID  `gorm:"type:uuid;" json:"userId"`
	TeamID *uuid.UUID `gorm:"type:uuid;" json:"teamId"` // Team is optional
	Team   Team
}

func (project *Project) BeforeCreate(db *gorm.DB) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	project.ID = id
	return nil
}
