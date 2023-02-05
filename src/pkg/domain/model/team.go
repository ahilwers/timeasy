package model

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Name string
}

func (team *Team) BeforeCreate(db *gorm.DB) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	team.ID = id
	return nil
}
