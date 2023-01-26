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

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_projectHandler_GetProjectById(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)
	token, user := loginUser(t, router, userUsecase, model.User{
		Username: "user",
		Password: "password",
	})

	project := model.Project{
		Name:   "testproject",
		UserId: user.ID,
	}
	err := projectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, user.ID, projectFromService.UserId)
}

func Test_projectHandler_GetProjectByIdFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	router := SetupRouter(userHandler, projectHandler)
	token, _ := loginUser(t, router, userUsecase, model.User{
		Username: "user",
		Password: "password",
	})

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", projectId), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", projectId))
}

func Test_projectHandler_GetProjectByIdFailsIfItDoesNotBelongToUser(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	projectOwner := model.User{
		Username: "owner",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&projectOwner)

	router := SetupRouter(userHandler, projectHandler)
	token, _ := loginUser(t, router, userUsecase, model.User{
		Username: "user",
		Password: "password",
	})

	project := model.Project{
		Name:   "testproject",
		UserId: projectOwner.ID,
	}
	err = projectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", project.ID))
}

func Test_projectHandler_GetProjectByIdPassesIfBelongsToOtherUserAndUserIsAdmin(t *testing.T) {
	teardownTest := test.SetupTest(t)
	defer teardownTest(t)

	projectRepo := database.NewGormProjectRepository(test.DB)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)

	userHandler := NewUserHandler(userUsecase)
	projectHandler := NewProjectHandler(projectUsecase)

	projectOwner := model.User{
		Username: "owner",
		Password: "password",
	}
	_, err := userUsecase.AddUser(&projectOwner)

	router := SetupRouter(userHandler, projectHandler)
	token, _ := loginUser(t, router, userUsecase, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleAdmin, model.RoleUser},
	})

	project := model.Project{
		Name:   "testproject",
		UserId: projectOwner.ID,
	}
	err = projectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, projectOwner.ID, projectFromService.UserId)
}

func loginUser(t *testing.T, router *gin.Engine, userUsecase usecase.UserUsecase, user model.User) (string, *model.User) {
	username := user.Username
	password := user.Password
	_, err := userUsecase.AddUser(&user)
	assert.Nil(t, err)

	token, err := Login(router, username, password)
	assert.Nil(t, err)

	return token, &user
}
