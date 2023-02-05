package model

import (
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;"`
	Username string
	Password string
	Roles    RoleList `gorm:"type:VARCHAR(255)"` //store the roles in a string field
	Teams    []UserTeamAssignment
}

type UserTeamAssignment struct {
	gorm.Model
	UserID uuid.UUID
	TeamID uuid.UUID
	Team   Team
	Roles  RoleList `gorm:"type:VARCHAR(255)"` //store the team roles in a string field
}

func (user *User) BeforeCreate(db *gorm.DB) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	user.ID = id
	// If no group is provided add the user role
	if len(user.Roles) == 0 {
		user.Roles = append(user.Roles, RoleUser)
	}
	return nil
}
