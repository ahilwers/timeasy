package usecase

import (
	"fmt"
	"strings"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, err, fmt.Sprintf("error adding user: %v", err))

	users, err := userRepo.GetAllUsers()
	assert.Nil(t, err, fmt.Sprintf("error getting users: %v", err))
	assert.Equal(t, 1, len(users))
	userFromDb := users[0]
	assert.Equal(t, user.Username, userFromDb.Username)
	// After Adding the user should have the Role "USER"
	assert.Equal(t, 1, len(userFromDb.Roles), "the user should have a role")
	assert.Equal(t, model.RoleUser, userFromDb.Roles[0])
}

func Test_userUsecase_AddUserWithMultipleRoles(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)
	user := model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleAdmin, model.RoleUser},
	}
	_, err := userUsecase.AddUser(&user)
	assert.Nil(t, err, fmt.Sprintf("error adding user: %v", err))
	users, err := userRepo.GetAllUsers()
	assert.Nil(t, err, fmt.Sprintf("error getting users: %v", err))
	assert.Equal(t, 1, len(users))
	userFromDb := users[0]
	assert.Equal(t, user.Username, userFromDb.Username)
	// After Adding the user should have all the roles we provided
	assert.Equal(t, 2, len(userFromDb.Roles), "the user should have two roles")
	assert.Equal(t, model.RoleAdmin, userFromDb.Roles[0])
	assert.Equal(t, model.RoleUser, userFromDb.Roles[1])
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
	assert.NotNil(t, err, "adding a user without password should cause an error")
	assert.Equal(t, "password must not be empty", err.Error())
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
	assert.NotNil(t, err, "adding a user without password should cause an error")
	assert.Equal(t, "password must not be empty", err.Error())
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
	assert.NotNil(t, err, "adding a user without username should cause an error")
	assert.Equal(t, "username must not be empty", err.Error())
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
	assert.NotNil(t, err, "adding a user without username should cause an error")
	assert.Equal(t, "username must not be empty", err.Error())
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
	assert.Nil(t, err)
	users, err := userRepo.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	userFromDb := users[0]
	assert.Equal(t, user.Username, userFromDb.Username)
	// Password should not be the same as the one from the created user.
	assert.NotEqual(t, "password", userFromDb.Password, fmt.Sprintf("the password does not seem to be encrypted: %v",
		userFromDb.Password))
	assert.NotEqual(t, 0, strings.TrimSpace(userFromDb.Password), "the password of the saved user should not be empty")
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
	assert.Nil(t, err)

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = userUsecase.AddUser(&user2)
	assert.Nil(t, err)
	user1FromDb, err := userUsecase.GetUserById(user1.ID)
	assert.Nil(t, err)
	assert.Equal(t, user1.Username, user1FromDb.Username)

	user2FromDb, err := userUsecase.GetUserById(user2.ID)
	assert.Nil(t, err)
	assert.Equal(t, user2.Username, user2FromDb.Username)
}

func Test_userUsecase_GetUserByIdThrowsErrorIfUserNotFound(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = userUsecase.GetUserById(id)
	assert.NotNil(t, err)
}

func Test_userUsecase_GetUserByName(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	user1 := model.User{
		Username: "user1",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user1)
	assert.Nil(t, err)

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = userUsecase.AddUser(&user2)
	assert.Nil(t, err)

	user1FromDb, err := userUsecase.GetUserByName(user1.Username)
	assert.Nil(t, err)
	assert.Equal(t, user1.ID, user1FromDb.ID)

	user2FromDb, err := userUsecase.GetUserByName(user2.Username)
	assert.Nil(t, err)
	assert.Equal(t, user2.ID, user2FromDb.ID)
}

func Test_userUsecase_GetUserByNameThrowsErrorIfUserNotFound(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	_, err := userUsecase.GetUserByName("notExistingUser")
	assert.NotNil(t, err)
}

func Test_userUsecase_GetAllUsers(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	user1 := model.User{
		Username: "user1",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user1)
	assert.Nil(t, err)

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = userUsecase.AddUser(&user2)
	assert.Nil(t, err)

	usersFromDb, err := userUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(usersFromDb))
	for i, user := range usersFromDb {
		assert.Equal(t, fmt.Sprintf("user%v", i+1), user.Username)
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
	assert.Nil(t, err)
	user.Username = "updatedUser"
	err = userUsecase.UpdateUser(&user)
	assert.Nil(t, err)
	users, err := userRepo.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	updatedUser := users[0]
	assert.Equal(t, "updatedUser", updatedUser.Username)
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
	assert.Nil(t, err)

	originalHashedPassword := user.Password
	user.Username = "updatedUser"
	user.Password = "updatedPassword"
	err = userUsecase.UpdateUser(&user)
	assert.Nil(t, err)
	users, err := userRepo.GetAllUsers()
	assert.Nil(t, err)
	updatedUser := users[0]
	assert.Equal(t, originalHashedPassword, updatedUser.Password, "the password should not have been updated")
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
	assert.NotNil(t, err)
}

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
	assert.Nil(t, err)

	oldPassword := user.Password
	err = userUsecase.UpdateUserPassword(user.ID, "newPassword")
	assert.Nil(t, err)

	users, err := userRepo.GetAllUsers()
	assert.Nil(t, err)

	assert.Equal(t, 1, len(users))

	updatedUser := users[0]
	assert.NotEqual(t, oldPassword, updatedUser.Password, "the password was not updated")
	assert.NotEqual(t, "newPassword", updatedUser.Password, "the password was not encrypted")
}

func Test_updatingUserPasswordFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	notExistingId, _ := uuid.NewV4()
	err := userUsecase.UpdateUserPassword(notExistingId, "newPassword")
	assert.NotNil(t, err)
}

func Test_deleteUser(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)
	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	assert.Nil(t, err)

	err = userUsecase.DeleteUser(user.ID)
	assert.Nil(t, err)

	userList, err := userUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(userList))
}

func Test_deleteUserFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)
	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := NewUserUsecase(userRepo)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = userUsecase.DeleteUser(notExistingId)
	assert.NotNil(t, err)
}
