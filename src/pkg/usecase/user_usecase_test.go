package usecase

import (
	"strings"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"
)

func Test_userUsecase_AddUser(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	if err != nil {
		t.Errorf("error adding the user: %v", err)
	}
	users, err := userRepo.GetAllUsers()
	if err != nil {
		t.Errorf("error getting users from database: %v", err)
	}
	if len(users) == 0 {
		t.Error("user dataset was not created")
	}
	userFromDb := users[0]
	if userFromDb.Username != user.Username {
		t.Errorf("username wrong - expected: %v, actual: %v", user.Username, userFromDb.Username)
	}
}

func Test_userUseCase_AddingUserFailsIfPasswordEmpty(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "user",
	}
	_, err := userUsecase.AddUser(&user)
	if err == nil {
		t.Errorf("adding a user without password should cause an error")
	} else if err.Error() != "password must not be empty" {
		t.Errorf("wrong error was thrown: %v", err)
	}
}

func Test_userUseCase_AddingUserFailsIfPasswordOnlyContainsSpaces(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "user",
		Password: "  ",
	}
	_, err := userUsecase.AddUser(&user)
	if err == nil {
		t.Errorf("adding a user without password should cause an error")
	} else if err.Error() != "password must not be empty" {
		t.Errorf("wrong error was thrown: %v", err)
	}
}

func Test_userUseCase_AddingUserFailsIfUsernameEmpty(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	if err == nil {
		t.Errorf("adding a user without username should cause an error")
	} else if err.Error() != "username must not be empty" {
		t.Errorf("wrong error was thrown: %v", err)
	}
}

func Test_userUseCase_AddingUserFailsIfUsernameOnlyContainsSpaces(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "  ",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	if err == nil {
		t.Errorf("adding a user without username should cause an error")
	} else if err.Error() != "username must not be empty" {
		t.Errorf("wrong error was thrown: %v", err)
	}
}

func Test_userUsecase_IsPasswordEncryptedWhenAddingUser(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	if err != nil {
		t.Errorf("error adding the user: %v", err)
	}
	users, err := userRepo.GetAllUsers()
	if err != nil {
		t.Errorf("error getting users from database: %v", err)
	}
	if len(users) == 0 {
		t.Error("user dataset was not created")
	}
	userFromDb := users[0]
	if userFromDb.Username != user.Username {
		t.Errorf("username wrong - expected: %v, actual: %v", user.Username, userFromDb.Username)
	}
	// Password should not be the same as the one from the created user.
	if userFromDb.Password == "password" {
		t.Errorf("the password doesn't seem to be encrypted: %v", userFromDb.Password)
	}
	if len(strings.TrimSpace(userFromDb.Password)) == 0 {
		t.Error("the password of the saved user should not be empty")
	}
}
