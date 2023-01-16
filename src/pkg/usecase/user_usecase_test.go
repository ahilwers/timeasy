package usecase

import (
	"strings"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/gofrs/uuid"
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

func Test_userUsecase_GetUserById(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	user1 := model.User{
		Username: "user1",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user1)
	if err != nil {
		t.Errorf("error adding the user: %v", err)
	}

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = userUsecase.AddUser(&user2)
	if err != nil {
		t.Errorf("error adding the user: %v", err)
	}

	user1FromDb, err := userUsecase.GetUserById(user1.ID)
	if err != nil {
		t.Errorf("error getting user from database: %v", err)
	}
	if user1FromDb.Username != user1.Username {
		t.Errorf("username of found user doesn't match the searched user - searched: %v, actual: %v", user1.Username, user1FromDb.Username)
	}

	user2FromDb, err := userUsecase.GetUserById(user2.ID)
	if err != nil {
		t.Errorf("error getting user from database: %v", err)
	}
	if user2FromDb.Username != user2.Username {
		t.Errorf("username of found user doesn't match the searched user - searched: %v, actual: %v", user1.Username, user1FromDb.Username)
	}
}

func Test_userUsecase_GetUserByIdThrowsErrorIfUserNotFound(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	id, err := uuid.NewV4()
	if err != nil {
		t.Errorf("could not generate uuid: %v", err)
	}
	_, err = userUsecase.GetUserById(id)
	if err == nil {
		t.Error("there should have been an error as the user does not exist")
	}
}

func Test_userUsecase_UpdateUser(t *testing.T) {
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
	user.Username = "updatedUser"
	err = userUsecase.UpdateUser(&user)
	if err != nil {
		t.Errorf("error updating user: %v", err)
	}
	users, err := userRepo.GetAllUsers()
	if err != nil {
		t.Errorf("error getting users from database: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("there should only be 1 user in the database - actual count: %v", len(users))
	}
	updatedUser := users[0]
	if updatedUser.Username != "updatedUser" {
		t.Errorf("the user was not updated correctly - username is %v", updatedUser.Username)
	}
}

func Test_userUsecase_PasswordShouldNotBeChangedWhenUpdatingUser(t *testing.T) {
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
	originalHashedPassword := user.Password
	user.Username = "updatedUser"
	user.Password = "updatedPassword"
	err = userUsecase.UpdateUser(&user)
	if err != nil {
		t.Errorf("error updating user: %v", err)
	}
	users, err := userRepo.GetAllUsers()
	updatedUser := users[0]
	if updatedUser.Password != originalHashedPassword {
		t.Errorf("the password should not have been updated - expected: %v, actual: %v", originalHashedPassword, updatedUser.Password)
	}
}

func Test_userUsecase_UpdateUserFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	err := userUsecase.UpdateUser(&user)
	if err == nil {
		t.Errorf("an error was expected because the user does not exist")
	}
}

// Todo: Add tests for updating the password of a user.
func Test_userUsecase_UpdateUserPassword(t *testing.T) {
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
	oldPassword := user.Password
	err = userUsecase.UpdateUserPassword(user.ID, "newPassword")
	if err != nil {
		t.Errorf("error updating user password: %v", err)
	}
	users, err := userRepo.GetAllUsers()
	if err != nil {
		t.Errorf("error getting users from database: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("there should only be 1 user in the database - actual count: %v", len(users))
	}
	updatedUser := users[0]
	if updatedUser.Password == oldPassword {
		t.Errorf("the user password was not updated - password is %v", updatedUser.Password)
	}
	if updatedUser.Password == "newPassword" {
		t.Errorf("the user password was not encrypted - password is still %v", updatedUser.Password)
	}
}

func Test_updatingUserPasswordFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	notExistingId, _ := uuid.NewV4()
	err := userUsecase.UpdateUserPassword(notExistingId, "newPassword")
	if err == nil {
		t.Error("updating a password for a non existing user should have thrown an error.")
	}
}
