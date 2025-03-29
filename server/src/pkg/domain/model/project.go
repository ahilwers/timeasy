package model

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;" json:"id"`
	Name              string     `json:"name"`
	UserId            uuid.UUID  `gorm:"type:uuid;" json:"userId"`
	TeamID            *uuid.UUID `gorm:"type:uuid;" json:"teamId"` // Team is optional
	Team              Team
	Color             string          `gorm:"default:#1E90FF;" json:"color,omitempty"`
	Deadline          DateOnly        `gorm:"type:date;" json:"deadline,omitempty"`
	HourlyRate        decimal.Decimal `gorm:"type:numeric(10,2);default:0.00;" json:"hourlyRate,omitempty"`
	TimeBudgetInHours int             `gorm:"default:0;" json:"timeBudget,omitempty"`
}

func (project *Project) BeforeCreate(db *gorm.DB) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	project.ID = id
	return nil
}
