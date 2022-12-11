package usecase

import (
	"fmt"
	"strings"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	GetUserById(id uuid.UUID) (*model.User, error)
	AddUser(user *model.User) (*model.User, error)
	UpdateUser(user *model.User) error
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: repo,
	}
}

func (uu *userUsecase) GetUserById(id uuid.UUID) (*model.User, error) {
	return uu.userRepo.GetUserById(id)
}

func (uu *userUsecase) AddUser(user *model.User) (*model.User, error) {
	err := uu.checkUserData(user)
	if err != nil {
		return user, err
	}
	hashedPassword, err := encryptPassword(user.Password)
	if err != nil {
		return user, err
	}
	user.Password = hashedPassword
	return uu.userRepo.AddUser(user)
}

func (uu *userUsecase) UpdateUser(user *model.User) error {
	err := uu.checkUserData(user)
	if err != nil {
		return err
	}
	userFromDb, err := uu.GetUserById(user.ID)
	if err != nil {
		return err
	}
	user.Password = userFromDb.Password
	return uu.userRepo.UpdateUser(user)
}

func (uu *userUsecase) checkUserData(user *model.User) error {
	if len(strings.TrimSpace(user.Username)) == 0 {
		return fmt.Errorf("username must not be empty")
	}
	if len(strings.TrimSpace(user.Password)) == 0 {
		return fmt.Errorf("password must not be empty")
	}
	return nil
}

func encryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return password, err
	}
	return string(hashedPassword), nil
}
