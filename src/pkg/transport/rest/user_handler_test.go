package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_userHandler_GetUserById(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	TestUserUsecase.AddUser(&user)

	token, err := Login("user", "password")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_userHandler_GetUserByIdFailsIfNotLoggedIn(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.ID), nil)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func Test_userHandler_GetUserByIdFailsIfOtherUserAndNotAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	loginUser := model.User{
		Username: "user1",
		Password: "password1",
	}
	_, err := TestUserUsecase.AddUser(&loginUser)
	assert.Nil(t, err)
	addedUser := model.User{
		Username: "user2",
		Password: "password2",
	}
	_, err = TestUserUsecase.AddUser(&addedUser)
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", addedUser.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_userHandler_GetOtherUserByIdPassesIfUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	loginUser := model.User{
		Username: "user1",
		Password: "password1",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	}
	_, err := TestUserUsecase.AddUser(&loginUser)
	assert.Nil(t, err)
	addedUser := model.User{
		Username: "user2",
		Password: "password2",
	}
	_, err = TestUserUsecase.AddUser(&addedUser)
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", addedUser.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_userHandler_GetUserByIdFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	loginUser := model.User{
		Username: "user1",
		Password: "password1",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	}
	_, err := TestUserUsecase.AddUser(&loginUser)
	assert.Nil(t, err)

	notExistingUser := model.User{
		Username: "user2",
		Password: "password2",
	}

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", notExistingUser.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func Test_userHandler_GetUserList(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("loginuser", "loginpassword", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	users, err := addUsers(3)
	assert.Nil(t, err)

	token, err := Login("loginuser", "loginpassword")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var usersFromService []model.User
	json.Unmarshal(w.Body.Bytes(), &usersFromService)
	assert.Equal(t, len(users)+1, len(usersFromService))
	for i, user := range usersFromService {
		if i == 0 {
			assert.Equal(t, "loginuser", user.Username)
		} else {
			assert.Equal(t, fmt.Sprintf("user%v", i), user.Username)
		}
	}
}

func Test_userHandler_GetUserListFailsIfUserIsNotAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("loginuser", "loginpassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	_, err = addUsers(3)
	assert.Nil(t, err)

	token, err := Login("loginuser", "loginpassword")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func Test_userHandler_Signup(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"username\": \"%v\", \"password\": \"%v\"}", "user1", "password1"))
	req, err := http.NewRequest("POST", "/api/v1/signup", reader)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	usersFromDb, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(usersFromDb))
	assert.Equal(t, "user1", usersFromDb[0].Username)
}

func Test_userHandler_SignupFailsWithoutUsername(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"username\": \"%v\", \"password\": \"%v\"}", "", "password"))
	req, err := http.NewRequest("POST", "/api/v1/signup", reader)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	usersFromDb, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(usersFromDb))
}

func Test_userHandler_SignupFailsWithoutPassword(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"username\": \"%v\", \"password\": \"%v\"}", "user1", ""))
	req, err := http.NewRequest("POST", "/api/v1/signup", reader)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	usersFromDb, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(usersFromDb))
}

func Test_userHandler_SignupFailsIfUsernameExists(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)
	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser})

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"username\": \"%v\", \"password\": \"%v\"}", "user1", "password"))
	req, err := http.NewRequest("POST", "/api/v1/signup", reader)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 409, w.Code)

	AssertErrorMessageEquals(t, w.Body.Bytes(), "a user with the same name already exists")

	usersFromDb, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(usersFromDb))
}

func Test_userHandler_UpdateUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user, err := addUser("user1", "password1", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	user.Username = "updatedUser"
	userJson, err := json.Marshal(user)
	assert.Nil(t, err)
	reader := bytes.NewReader(userJson)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v", user.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	usersFromDb, err := TestUserUsecase.GetAllUsers()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(usersFromDb))
	assert.Equal(t, user.ID, usersFromDb[0].ID)
	assert.Equal(t, "updatedUser", usersFromDb[0].Username)
}

func Test_userHandler_UpdateUserFailsIfUserUpdatesAnotherUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	otherUser.Username = "updatedUser"
	userJson, err := json.Marshal(otherUser)
	assert.Nil(t, err)
	reader := bytes.NewReader(userJson)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	// The other user should not be updated:
	otherUserFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, "otherUser", otherUserFromDb.Username)
}

func Test_userHandler_UpdateUserPassesIfUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	otherUser.Username = "updatedUser"
	userJson, err := json.Marshal(otherUser)
	assert.Nil(t, err)
	reader := bytes.NewReader(userJson)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	otherUserFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, "updatedUser", otherUserFromDb.Username)
}

func Test_userHandler_UpdateUserFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	otherUser := model.User{
		ID:       userId,
		Username: "otherUser",
		Password: "otherPassword",
	}
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	otherUser.Username = "updatedUser"
	userJson, err := json.Marshal(otherUser)
	assert.Nil(t, err)
	reader := bytes.NewReader(userJson)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func Test_userHandler_UpdateUserFailsIfUsernameIsEmpty(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	otherUser.Username = ""
	userJson, err := json.Marshal(otherUser)
	assert.Nil(t, err)
	reader := bytes.NewReader(userJson)
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	AssertErrorMessageEquals(t, w.Body.Bytes(), "username must not be empty")

	otherUserFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, "otherUser", otherUserFromDb.Username)
}

func Test_userHandler_UpdateUserPassword(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user, err := addUser("user1", "password1", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"password\": \"%v\"}", "newPassword"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v/password", user.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	userFromDb, err := TestUserUsecase.GetUserById(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, user.Username, userFromDb.Username)
	err = TestUserUsecase.VerifyPassword("newPassword", userFromDb.Password)
	assert.Nil(t, err)
}

func Test_userHandler_UpdateUserPasswordFailsIfNotSameUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"password\": \"%v\"}", "newPassword"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v/password", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	userFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, otherUser.Username, userFromDb.Username)
	// Password should still be the same:
	err = TestUserUsecase.VerifyPassword("otherPassword", userFromDb.Password)
	assert.Nil(t, err)
}

func Test_userHandler_UpdateUserPasswordPassesIfUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"password\": \"%v\"}", "newPassword"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v/password", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	userFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, otherUser.Username, userFromDb.Username)
	err = TestUserUsecase.VerifyPassword("newPassword", userFromDb.Password)
	assert.Nil(t, err)
}

func Test_userHandler_UpdateUserPasswordFailsIfPasswordEmpty(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"password\": \"%v\"}", ""))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v/password", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)

	userFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, otherUser.Username, userFromDb.Username)
	// Password should still be the same:
	err = TestUserUsecase.VerifyPassword("otherPassword", userFromDb.Password)
	assert.Nil(t, err)
}

func Test_userHandler_UpdateUserRoles(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)
	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader("{\"roles\": [\"USER\", \"ADMIN\"]}")
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v/roles", otherUser.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	userFromDb, err := TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(userFromDb.Roles))
	assert.Equal(t, model.RoleUser, userFromDb.Roles[0])
	assert.Equal(t, model.RoleAdmin, userFromDb.Roles[1])
}

func Test_userHandler_UpdateUserRolesFailsIfNotAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	user, err := addUser("user1", "password1", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader("{\"roles\": [\"USER\", \"ADMIN\"]}")
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%v/roles", user.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	userFromDb, err := TestUserUsecase.GetUserById(user.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(userFromDb.Roles))
	assert.Equal(t, model.RoleUser, userFromDb.Roles[0])
}

func Test_userHandler_DeleteUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)
	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%v", otherUser.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	_, err = TestUserUsecase.GetUserById(otherUser.ID)
	assert.NotNil(t, err)
}

func Test_userHandler_DeleteUserFailsIfNotAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser})
	assert.Nil(t, err)
	otherUser, err := addUser("otherUser", "otherPassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%v", otherUser.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)

	_, err = TestUserUsecase.GetUserById(otherUser.ID)
	assert.Nil(t, err)
}

func Test_userHandler_DeleteUserFailsIfNotExists(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	_, err := addUser("user1", "password1", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	token, err := Login("user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	notExistingId, err := uuid.NewV4()
	assert.Nil(t, err)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%v", notExistingId), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func addUsers(count int) ([]model.User, error) {
	var users []model.User
	for i := 0; i < count; i++ {
		user := model.User{
			Username: fmt.Sprintf("user%v", i+1),
			Password: fmt.Sprintf("password%v", i+1),
		}
		_, err := TestUserUsecase.AddUser(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func addUser(name string, pasword string, roles model.RoleList) (*model.User, error) {
	user := createUser(name, pasword, roles)
	_, err := TestUserUsecase.AddUser(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func createUser(name string, pasword string, roles model.RoleList) model.User {
	return model.User{
		Username: name,
		Password: pasword,
		Roles:    roles,
	}
}
