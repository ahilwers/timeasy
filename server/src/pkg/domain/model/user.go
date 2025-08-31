package model

import (
	"github.com/gofrs/uuid"
)

type User struct {
	ID          uuid.UUID         `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	FirstName   string            `json:"firstName"`
	LastName    string            `json:"lastName"`
	DisplayName string            `json:"displayName"`
	Language    *string           `json:"language,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type UserProfileUpdateRequest struct {
	FirstName   *string           `json:"firstName,omitempty"`
	LastName    *string           `json:"lastName,omitempty"`
	Email       *string           `json:"email,omitempty"`
	Language    *string           `json:"language,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}