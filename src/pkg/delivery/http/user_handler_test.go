package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"
	"timeasy-server/pkg/usecase"
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
	if err != nil {
		t.Errorf("error logging in: %v", err)
	}
	fmt.Printf("Token: %v", token)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("error getting a user - return code is %v", w.Code)
	}
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
	userUsecase.AddUser(&user)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.ID), nil)
	router.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("getting a user without login should return code 401 - actual code: %v", w.Code)
	}
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
	userUsecase.AddUser(&loginUser)
	addedUser := model.User{
		Username: "user2",
		Password: "password2",
	}
	userUsecase.AddUser(&addedUser)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user1", "password1")
	if err != nil {
		t.Errorf("error logging in: %v", err)
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", addedUser.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("a user should not be able to get the info of another user - returned code: %v", w.Code)
	}
}

func Test_userHandler_GetUserByIdPassesIfUserIsAdmin(t *testing.T) {
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
	userUsecase.AddUser(&loginUser)
	addedUser := model.User{
		Username: "user2",
		Password: "password2",
	}
	userUsecase.AddUser(&addedUser)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user1", "password1")
	if err != nil {
		t.Errorf("error logging in: %v", err)
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", addedUser.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("an admin should be able to get the data of another user - returned code: %v", w.Code)
	}
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
	userUsecase.AddUser(&loginUser)
	notExistingUser := model.User{
		Username: "user2",
		Password: "password2",
	}

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)

	token, err := Login(router, "user1", "password1")
	if err != nil {
		t.Errorf("error logging in: %v", err)
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", notExistingUser.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("user should have not been found - returned code: %v", w.Code)
	}
}
