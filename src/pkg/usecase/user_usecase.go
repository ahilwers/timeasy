package usecase

import (
	"fmt"
	"strings"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"golang.org/x/crypto/bcrypt"
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
	hashedPassword, err := encryptPassword(user.Password)
	if err != nil {
		return user, err
	}
	user.Password = hashedPassword
	return uu.userRepo.AddUser(user)
}

func encryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return password, err
	}
	return string(hashedPassword), nil
}
