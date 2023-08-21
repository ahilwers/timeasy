package model

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Name   string
	UserId uuid.UUID  `gorm:"type:uuid;"`
	TeamID *uuid.UUID `gorm:"type:uuid;"` // Team is optional
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
