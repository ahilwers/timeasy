package usecase

import (
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/test"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
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
	SyncUsecase      SyncUsecase
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
	teamRepo := postgresql.NewPostgreSQLTeamRepository(test.Database.DB)
	u.TeamUsecase = NewTeamUsecase(teamRepo)

	projectRepo := postgresql.NewPostgreSQLProjectRepository(test.Database.DB, teamRepo)
	u.ProjectUsecase = NewProjectUsecase(projectRepo, u.TeamUsecase)

	timeEntryRepo := postgresql.NewPostgreSQLTimeEntryRepository(test.Database.DB)
	u.TimeEntryUsecase = NewTimeEntryUsecase(timeEntryRepo, u.ProjectUsecase)

	//syncRepo := database.NewGormSyncRepository(test.DB)
	//u.SyncUsecase = NewSyncUsecase(syncRepo)
}

func GetTestUserId(t *testing.T) uuid.UUID {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	return userId
}
