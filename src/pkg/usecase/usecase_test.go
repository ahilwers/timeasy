package usecase

import (
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/test"
)

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}

type UsecaseTest struct {
	ProjectUsecase   ProjectUsecase
	TimeEntryUsecase TimeEntryUsecase
	TeamUsecase      TeamUsecase
}

func NewUsecaseTest() *UsecaseTest {
	return &UsecaseTest{}
}

func (u *UsecaseTest) SetupTest(tb testing.TB) func(tb testing.TB) {
	tearDownTest := test.SetupTest(tb)
	u.initUsecases()
	return tearDownTest
}

func (u *UsecaseTest) initUsecases() {
	projectRepo := database.NewGormProjectRepository(test.DB)
	u.ProjectUsecase = NewProjectUsecase(projectRepo)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	u.TimeEntryUsecase = NewTimeEntryUsecase(timeEntryRepo, u.ProjectUsecase)

	teamRepo := database.NewGormTeamRepository(test.DB)
	u.TeamUsecase = NewTeamUsecase(teamRepo)
}
