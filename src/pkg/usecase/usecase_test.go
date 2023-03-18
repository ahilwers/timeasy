package usecase

import (
	"fmt"
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/test"

	"github.com/stretchr/testify/assert"
)

var TestProjectUsecase ProjectUsecase
var TestTimeEntryUsecase TimeEntryUsecase
var TestTeamUsecase TeamUsecase

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}

func SetupTest(tb testing.TB) func(tb testing.TB) {
	tearDownTest := test.SetupTest(tb)
	initUsecases()
	return tearDownTest
}

func initUsecases() {
	projectRepo := database.NewGormProjectRepository(test.DB)
	TestProjectUsecase = NewProjectUsecase(projectRepo)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	TestTimeEntryUsecase = NewTimeEntryUsecase(timeEntryRepo, TestProjectUsecase)

	teamRepo := database.NewGormTeamRepository(test.DB)
	TestTeamUsecase = NewTeamUsecase(teamRepo)
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
