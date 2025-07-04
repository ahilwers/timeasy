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
	Roles  RoleList
}
