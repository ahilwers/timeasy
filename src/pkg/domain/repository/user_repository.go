package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type UserRepository interface {
	AddUser(user *model.User) (*model.User, error)
	UpdateUser(user *model.User) error
	GetUserById(id uuid.UUID) (*model.User, error)
	GetUserByName(username string) (*model.User, error)
	GetAllUsers() ([]model.User, error)
}
