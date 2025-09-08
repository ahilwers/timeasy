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

