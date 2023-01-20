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
	GetUserByName(username string) (*model.User, error)
	GetAllUsers() ([]model.User, error)
	AddUser(user *model.User) (*model.User, error)
	// Updates a user
	// Note: This will not update the password - use UpdateUserPassword if you want to update the password.
	UpdateUser(user *model.User) error
	// Updates the password of an existing user with the specified id.
	UpdateUserPassword(id uuid.UUID, newPassword string) error
	// Checks if the given password is equal to the hashed password.
	VerifyPassword(password, hashedPassword string) error
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

func (uu *userUsecase) GetUserByName(username string) (*model.User, error) {
	return uu.userRepo.GetUserByName(username)
}

func (uu *userUsecase) GetAllUsers() ([]model.User, error) {
	return uu.userRepo.GetAllUsers()
}

func (uu *userUsecase) AddUser(user *model.User) (*model.User, error) {
	err := uu.checkUserData(user)
	if err != nil {
		return user, err
	}
	hashedPassword, err := uu.encryptPassword(user.Password)
	if err != nil {
		return user, fmt.Errorf("could not encrypt password: %v", err)
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

func (uu *userUsecase) UpdateUserPassword(id uuid.UUID, newPassword string) error {
	hashedPassword, err := uu.encryptPassword(newPassword)
	if err != nil {
		return fmt.Errorf("could not encrypt password: %v", err)
	}
	user, err := uu.GetUserById(id)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	uu.userRepo.UpdateUser(user)
	return nil
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

func (uu *userUsecase) encryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return password, err
	}
	return string(hashedPassword), nil
}

func (uu *userUsecase) VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
