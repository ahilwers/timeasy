package http

import (
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

func Test_projectHandler_GetProjectById(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	project := model.Project{
		Name:   "testproject",
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, user.ID, projectFromService.UserId)
}

func Test_projectHandler_GetProjectByIdFailsIfProjectDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", projectId), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", projectId))
}

func Test_projectHandler_GetProjectByIdFailsIfItDoesNotBelongToUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	projectOwner := model.User{
		Username: "owner",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&projectOwner)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	project := model.Project{
		Name:   "testproject",
		UserId: projectOwner.ID,
	}
	err = TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), fmt.Sprintf("project with id %v not found", project.ID))
}

func Test_projectHandler_GetProjectByIdPassesIfBelongsToOtherUserAndUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	projectOwner := model.User{
		Username: "owner",
		Password: "password",
	}
	_, err := TestUserUsecase.AddUser(&projectOwner)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleAdmin, model.RoleUser},
	})

	project := model.Project{
		Name:   "testproject",
		UserId: projectOwner.ID,
	}
	err = TestProjectUsecase.AddProject(&project)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectFromService model.Project
	json.Unmarshal(w.Body.Bytes(), &projectFromService)
	assert.Equal(t, project.Name, projectFromService.Name)
	assert.Equal(t, projectOwner.ID, projectFromService.UserId)
}

func Test_projectHandler_GetAllProjects(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	addProjects(t, 3, user)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectsFromService []model.Project
	json.Unmarshal(w.Body.Bytes(), &projectsFromService)
	for i, project := range projectsFromService {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
		assert.Equal(t, user.ID, project.UserId)
	}
}

func Test_projectHandler_GetAllProjectsReturnsOnlyProjectsOfUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)
	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
	})

	addProjects(t, 3, user)
	otherUser := model.User{
		Username: "otherUser",
		Password: "otherPassword",
	}
	_, err := TestUserUsecase.AddUser(&otherUser)
	assert.Nil(t, err)
	addProjectsWithStartIndex(t, 4, 3, otherUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectsFromService []model.Project
	json.Unmarshal(w.Body.Bytes(), &projectsFromService)
	assert.Equal(t, 3, len(projectsFromService))
	for i, project := range projectsFromService {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
		assert.Equal(t, user.ID, project.UserId)
	}
}

func Test_projectHandler_GetAllProjectsReturnsAllProjectsIfUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)
	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})

	addProjects(t, 3, user)
	otherUser := model.User{
		Username: "otherUser",
		Password: "otherPassword",
	}
	_, err := TestUserUsecase.AddUser(&otherUser)
	assert.Nil(t, err)
	addProjectsWithStartIndex(t, 4, 3, otherUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects", nil)
	AddToken(req, token)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	var projectsFromService []model.Project
	json.Unmarshal(w.Body.Bytes(), &projectsFromService)
	assert.Equal(t, 6, len(projectsFromService))
	for i, project := range projectsFromService {
		assert.Equal(t, fmt.Sprintf("Project %v", i+1), project.Name)
	}
}

func Test_projectHandler_AddProject(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\"}", "project1"))
	req, err := http.NewRequest("POST", "/api/v1/projects", reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "project1", projectsFromDb[0].Name)
	assert.Equal(t, user.ID, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProject(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	project := addProject(t, "project", user)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\"}", "updatedProject"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.Equal(t, user.ID, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProjectFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\"}", "updatedProject"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", projectId), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectHandler_UpdateProjectFailsIfItBelongsToAnotherUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	projectOwner := model.User{
		Username: "owner",
		Password: "ownerpassword",
	}
	_, err := TestUserUsecase.AddUser(&projectOwner)
	assert.Nil(t, err)

	project := addProject(t, "project", projectOwner)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\"}", "updatedProject"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Code)
	AssertErrorMessageEquals(t, w.Body.Bytes(), "you are not allowed to update this project")

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "project", projectsFromDb[0].Name)
	assert.Equal(t, projectOwner.ID, projectsFromDb[0].UserId)
}

func Test_projectHandler_UpdateProjectSucceedsIfItBelongsToAnotherUserAndUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})

	projectOwner := model.User{
		Username: "owner",
		Password: "ownerpassword",
	}
	_, err := TestUserUsecase.AddUser(&projectOwner)
	assert.Nil(t, err)

	project := addProject(t, "project", projectOwner)

	w := httptest.NewRecorder()
	reader := strings.NewReader(fmt.Sprintf("{\"name\": \"%v\"}", "updatedProject"))
	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%v", project.ID), reader)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
	assert.Equal(t, "updatedProject", projectsFromDb[0].Name)
	assert.Equal(t, projectOwner.ID, projectsFromDb[0].UserId)
}

func Test_projectHandler_DeleteProject(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, user := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	project := addProject(t, "project", user)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func Test_projectHandler_DeleteProjectFailsIfItDoesNotExist(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	projectId, err := uuid.NewV4()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", projectId), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func Test_projectHandler_DeleteProjectFailsIfItBelongsToAnotherUser(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser},
	})

	projectOwner := model.User{
		Username: "owner",
		Password: "ownerpassword",
	}
	_, err := TestUserUsecase.AddUser(&projectOwner)

	project := addProject(t, "project", projectOwner)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(projectsFromDb))
}

func Test_projectHandler_DeleteProjectSucceedsIfItBelongsToAnotherUserAndUserIsAdmin(t *testing.T) {
	teardownTest := SetupTest(t)
	defer teardownTest(t)

	token, _ := loginUser(t, model.User{
		Username: "user",
		Password: "password",
		Roles:    model.RoleList{model.RoleUser, model.RoleAdmin},
	})

	projectOwner := model.User{
		Username: "owner",
		Password: "ownerpassword",
	}
	_, err := TestUserUsecase.AddUser(&projectOwner)

	project := addProject(t, "project", projectOwner)

	w := httptest.NewRecorder()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%v", project.ID), nil)
	AddToken(req, token)
	assert.Nil(t, err)
	TestRouter.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	projectsFromDb, err := TestProjectUsecase.GetAllProjects()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(projectsFromDb))
}

func loginUser(t *testing.T, user model.User) (string, model.User) {
	username := user.Username
	password := user.Password
	_, err := TestUserUsecase.AddUser(&user)
	assert.Nil(t, err)

	token, err := Login(username, password)
	assert.Nil(t, err)

	return token, user
}

func addProjects(t *testing.T, count int, user model.User) []model.Project {
	return addProjectsWithStartIndex(t, 1, count, user)
}

func addProjectsWithStartIndex(t *testing.T, startIndex int, count int, user model.User) []model.Project {
	var projects []model.Project
	for i := 0; i < count; i++ {
		project := addProject(t, fmt.Sprintf("Project %v", startIndex+i), user)
		projects = append(projects, project)
	}
	return projects
}

func addProject(t *testing.T, name string, user model.User) model.Project {
	prj := model.Project{
		Name:   name,
		UserId: user.ID,
	}
	err := TestProjectUsecase.AddProject(&prj)
	assert.Nil(t, err)
	return prj
}
