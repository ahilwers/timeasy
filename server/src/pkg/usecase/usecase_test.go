package usecase

import (
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/domain/repository"
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
	ChangelogRepo    repository.ChangelogRepository
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
	u.ChangelogRepo = postgresql.NewPostgreSQLChangelogRepository(test.Database.DB)
	teamRepo := postgresql.NewPostgreSQLTeamRepository(test.Database.DB)
	u.TeamUsecase = NewTeamUsecase(teamRepo, u.ChangelogRepo)

	projectRepo := postgresql.NewPostgreSQLProjectRepository(test.Database.DB, teamRepo)
	u.ProjectUsecase = NewProjectUsecase(projectRepo, u.TeamUsecase, u.ChangelogRepo)

	timeEntryRepo := postgresql.NewPostgreSQLTimeEntryRepository(test.Database.DB)
	u.TimeEntryUsecase = NewTimeEntryUsecase(timeEntryRepo, u.ProjectUsecase, u.ChangelogRepo)

	syncRepo := postgresql.NewPostgreSQLSyncRepository(test.Database.DB)
	u.SyncUsecase = NewSyncUsecase(syncRepo, u.ChangelogRepo, projectRepo, timeEntryRepo)
}

func GetTestUserId(t *testing.T) uuid.UUID {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	return userId
}

func GetTestClientId(t *testing.T) string {
	userId, err := uuid.NewV4()
	assert.Nil(t, err)
	return userId.String()
}
