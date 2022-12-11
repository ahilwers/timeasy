package usecase

import (
	"fmt"
	"strings"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
)

type UserUsecase interface {
	AddUser(user *model.User) (*model.User, error)
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: repo,
	}
}

func (uu *userUsecase) AddUser(user *model.User) (*model.User, error) {
	if len(strings.TrimSpace(user.Username)) == 0 {
		return nil, fmt.Errorf("username must not be empty")
	}
	if len(strings.TrimSpace(user.Password)) == 0 {
		return nil, fmt.Errorf("password must not be empty")
	}
	return uu.userRepo.AddUser(user)
}
