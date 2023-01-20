package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"
	"timeasy-server/pkg/usecase"

	"github.com/stretchr/testify/assert"
)

func Test_userHandler_GetUserById(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	userUsecase.AddUser(&user)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user", "password")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_userHandler_GetUserByIdFailsIfNotLoggedIn(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	user := model.User{
		Username: "user",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&user)
	assert.Nil(t, err)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.ID), nil)
	assert.Nil(t, err)
	router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func Test_userHandler_GetUserByIdFailsIfOtherUserAndNotAdmin(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	loginUser := model.User{
		Username: "user1",
		Password: "password1",
	}
	_, err := userUsecase.AddUser(&loginUser)
	assert.Nil(t, err)
	addedUser := model.User{
		Username: "user2",
		Password: "password2",
	}
	_, err = userUsecase.AddUser(&addedUser)
	assert.Nil(t, err)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", addedUser.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func Test_userHandler_GetOtherUserByIdPassesIfUserIsAdmin(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	loginUser := model.User{
		Username: "user1",
		Password: "password1",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	}
	_, err := userUsecase.AddUser(&loginUser)
	assert.Nil(t, err)
	addedUser := model.User{
		Username: "user2",
		Password: "password2",
	}
	_, err = userUsecase.AddUser(&addedUser)
	assert.Nil(t, err)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", addedUser.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func Test_userHandler_GetUserByIdFailsIfUserDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	loginUser := model.User{
		Username: "user1",
		Password: "password1",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	}
	_, err := userUsecase.AddUser(&loginUser)
	assert.Nil(t, err)

	notExistingUser := model.User{
		Username: "user2",
		Password: "password2",
	}

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user1", "password1")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", notExistingUser.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func Test_userHandler_GetUserList(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	_, err := addUser(userUsecase, "loginuser", "loginpassword", model.RoleList{model.RoleUser, model.RoleAdmin})
	assert.Nil(t, err)

	users, err := addUsers(userUsecase, 3)
	assert.Nil(t, err)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "loginuser", "loginpassword")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
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
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	_, err := addUser(userUsecase, "loginuser", "loginpassword", model.RoleList{model.RoleUser})
	assert.Nil(t, err)

	_, err = addUsers(userUsecase, 3)
	assert.Nil(t, err)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "loginuser", "loginpassword")
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func addUsers(userUsecase usecase.UserUsecase, count int) ([]model.User, error) {
	var users []model.User
	for i := 0; i < count; i++ {
		user := model.User{
			Username: fmt.Sprintf("user%v", i+1),
			Password: fmt.Sprintf("password%v", i+1),
		}
		_, err := userUsecase.AddUser(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func addUser(userUsecase usecase.UserUsecase, name string, pasword string, roles model.RoleList) (*model.User, error) {
	user := createUser(name, pasword, roles)
	_, err := userUsecase.AddUser(&user)
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
