package repository

import "timeasy-server/pkg/domain/model"

type UserRepository interface {
	AddUser(user *model.User) (*model.User, error)
	GetAllUsers() ([]model.User, error)
}
