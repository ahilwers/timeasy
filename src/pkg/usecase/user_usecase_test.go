package usecase

import (
	"fmt"
	"strings"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_userUsecase_AddUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err, fmt.Sprintf("error adding user: %v", err))

	users, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err, fmt.Sprintf("error getting users: %v", err))
	assert.Equal(t, 1, len(users))
	userFromDb := users[0]
	assert.Equal(t, user.Username, userFromDb.Username)
	// After Adding the user should have the Role "USER"
	assert.Equal(t, 1, len(userFromDb.Roles), "the user should have a role")
	assert.Equal(t, model.RoleUser, userFromDb.Roles[0])
}

func Test_userUsecase_AddUserWithMultipleRoles(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleAdmin, model.RoleUser},
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err, fmt.Sprintf("error adding user: %v", err))
	users, err := TestUserUsecase.GetAllUsers()
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
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.NotNil(t, err, "adding a user without password should cause an error")
	assert.Equal(t, "password must not be empty", err.Error())
}

func Test_userUseCase_AddingUserFailsIfPasswordOnlyContainsSpaces(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "  ",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.NotNil(t, err, "adding a user without password should cause an error")
	assert.Equal(t, "password must not be empty", err.Error())
}

func Test_userUseCase_AddingUserFailsIfUsernameEmpty(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.NotNil(t, err, "adding a user without username should cause an error")
	assert.Equal(t, "username must not be empty", err.Error())
}

func Test_userUseCase_AddingUserFailsIfUsernameOnlyContainsSpaces(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "  ",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.NotNil(t, err, "adding a user without username should cause an error")
	assert.Equal(t, "username must not be empty", err.Error())
}

func Test_userUsecase_IsPasswordEncryptedWhenAddingUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)
	users, err := TestUserUsecase.GetAllUsers()
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
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user1 := model.User{
		Username: "user1",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user1)
	assert.Nil(t, err)

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = TestUserUsecase.AddUser(&user2)
	assert.Nil(t, err)
	user1FromDb, err := TestUserUsecase.GetUserById(user1.ID)
	assert.Nil(t, err)
	assert.Equal(t, user1.Username, user1FromDb.Username)

	user2FromDb, err := TestUserUsecase.GetUserById(user2.ID)
	assert.Nil(t, err)
	assert.Equal(t, user2.Username, user2FromDb.Username)
}

func Test_userUsecase_GetUserByIdThrowsErrorIfUserNotFound(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	id, err := uuid.NewV4()
	assert.Nil(t, err)
	_, err = TestUserUsecase.GetUserById(id)
	assert.NotNil(t, err)
}

func Test_userUsecase_GetUserByName(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user1 := model.User{
		Username: "user1",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user1)
	assert.Nil(t, err)

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = TestUserUsecase.AddUser(&user2)
	assert.Nil(t, err)

	user1FromDb, err := TestUserUsecase.GetUserByName(user1.Username)
	assert.Nil(t, err)
	assert.Equal(t, user1.ID, user1FromDb.ID)

	user2FromDb, err := TestUserUsecase.GetUserByName(user2.Username)
	assert.Nil(t, err)
	assert.Equal(t, user2.ID, user2FromDb.ID)
}

func Test_userUsecase_GetUserByNameThrowsErrorIfUserNotFound(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := TestUserUsecase.GetUserByName("notExistingUser")
	assert.NotNil(t, err)
}

func Test_userUsecase_GetAllUsers(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user1 := model.User{
		Username: "user1",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user1)
	assert.Nil(t, err)

	user2 := model.User{
		Username: "user2",
		Password: "password",
	}
	_, err = TestUserUsecase.AddUser(&user2)
	assert.Nil(t, err)

	usersFromDb, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(usersFromDb))
	for i, user := range usersFromDb {
		assert.Equal(t, fmt.Sprintf("user%v", i+1), user.Username)
	}
}

func Test_userUsecase_UpdateUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)
	user.Username = "updatedUser"
	err = TestUserUsecase.UpdateUser(&user)
	assert.Nil(t, err)
	users, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	updatedUser := users[0]
	assert.Equal(t, "updatedUser", updatedUser.Username)
}

func Test_userUsecase_PasswordShouldNotBeChangedWhenUpdatingUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)

	originalHashedPassword := user.Password
	user.Username = "updatedUser"
	user.Password = "updatedPassword"
	err = TestUserUsecase.UpdateUser(&user)
	assert.Nil(t, err)
	users, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	updatedUser := users[0]
	assert.Equal(t, originalHashedPassword, updatedUser.Password, "the password should not have been updated")
}

func Test_userUsecase_UpdateUserFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	err := TestUserUsecase.UpdateUser(&user)
	assert.NotNil(t, err)
}

func Test_userUsecase_UpdateUserPassword(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)

	oldPassword := user.Password
	err = TestUserUsecase.UpdateUserPassword(user.ID, "newPassword")
	assert.Nil(t, err)

	users, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)

	assert.Equal(t, 1, len(users))

	updatedUser := users[0]
	assert.NotEqual(t, oldPassword, updatedUser.Password, "the password was not updated")
	assert.NotEqual(t, "newPassword", updatedUser.Password, "the password was not encrypted")
}

func Test_updatingUserPasswordFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	notExistingId, _ := uuid.NewV4()
	err := TestUserUsecase.UpdateUserPassword(notExistingId, "newPassword")
	assert.NotNil(t, err)
}

func Test_deleteUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)

	err = TestUserUsecase.DeleteUser(user.ID)
	assert.Nil(t, err)

	userList, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(userList))
}

func Test_deleteUserFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	err = TestUserUsecase.DeleteUser(notExistingId)
	assert.NotNil(t, err)
}
