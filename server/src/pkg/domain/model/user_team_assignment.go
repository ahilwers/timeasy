package model

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type UserTeamAssignment struct {
	gorm.Model
	UserID uuid.UUID
	TeamID uuid.UUID
	Team   Team
	Roles  RoleList `gorm:"type:VARCHAR(255)"` //store the team roles in a string field
}
