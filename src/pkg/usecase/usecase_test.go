package usecase

import (
	"fmt"
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}

func addUser(t *testing.T, userUsecase UserUsecase, username string, password string, roles model.RoleList) model.User {
	user := model.User{
		Username: username,
		Password: password,
		Roles:    roles,
	}
	_, err := userUsecase.AddUser(&user)
	assert.Nil(t, err)
	return user
}

func addProjects(t *testing.T, projectUsecase ProjectUsecase, count int, user model.User) []model.Project {
	return addProjectsWithStartIndex(t, projectUsecase, 1, count, user)
}

func addProjectsWithStartIndex(t *testing.T, projectUsecase ProjectUsecase, startIndex int, count int, user model.User) []model.Project {
	var projects []model.Project
	for i := 0; i < count; i++ {
		project := addProject(t, projectUsecase, fmt.Sprintf("Project %v", startIndex+i), user)
		projects = append(projects, project)
	}
	return projects
}

func addProject(t *testing.T, projectUsecase ProjectUsecase, name string, user model.User) model.Project {
	prj := model.Project{
		Name:   name,
		UserId: user.ID,
	}
	err := projectUsecase.AddProject(&prj)
	assert.Nil(t, err)
	return prj
}
