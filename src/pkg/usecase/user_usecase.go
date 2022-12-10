package usecase

import (
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
	return uu.userRepo.AddUser(user)
}
