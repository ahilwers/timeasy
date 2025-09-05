package model

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type Project struct {
	ID                 uuid.UUID          `json:"id"`
	Name               string             `json:"name"`
	UserId             uuid.UUID          `json:"userId"`
	TeamID             *uuid.UUID         `json:"teamId"` // Team is optional
	Team               Team
	Color              string             `json:"color,omitempty"`
	Deadline           DateOnly           `json:"deadline,omitempty"`
	HourlyRate         decimal.Decimal    `json:"hourlyRate,omitempty"`
	TimeBudget         int                `json:"timeBudget,omitempty"`
	IsActive           bool               `json:"isActive,omitempty"`
	Deleted            bool               `json:"deleted,omitempty"`
	ExternalConnection *ExternalConnection `json:"externalConnection,omitempty"`
}
