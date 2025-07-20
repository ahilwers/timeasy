package model

import (
	"github.com/gofrs/uuid"
)

type UserTeamAssignment struct {
	ID     uuid.UUID
	UserID uuid.UUID
	TeamID uuid.UUID
	Team   Team
	Roles  RoleList
}
