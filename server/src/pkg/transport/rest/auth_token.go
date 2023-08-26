package rest

import "github.com/gofrs/uuid"

type AuthToken interface {
	GetUserId() (uuid.UUID, error)
	GetRoles() ([]string, error)
	HasRole(role string) (bool, error)
}
