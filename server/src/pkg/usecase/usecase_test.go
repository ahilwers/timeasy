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
	ProjectUsecase      ProjectUsecase
	TimeEntryUsecase    TimeEntryUsecase
	TeamUsecase         TeamUsecase
	SyncUsecase         SyncUsecase
	ChangelogRepository repository.ChangelogRepository
	ProjectRepository   repository.ProjectRepository
	TimeEntryRepository repository.TimeEntryRepository
	TeamRepository      repository.TeamRepository
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
	u.ChangelogRepository = postgresql.NewPostgreSQLChangelogRepository(test.Database.DB)
	u.TeamRepository = postgresql.NewPostgreSQLTeamRepository(test.Database.DB)
	u.TeamUsecase = NewTeamUsecase(u.TeamRepository, u.ChangelogRepository)

	u.ProjectRepository = postgresql.NewPostgreSQLProjectRepository(test.Database.DB, u.TeamRepository)
	u.ProjectUsecase = NewProjectUsecase(u.ProjectRepository, u.TeamUsecase, u.ChangelogRepository)

	u.TimeEntryRepository = postgresql.NewPostgreSQLTimeEntryRepository(test.Database.DB)
	u.TimeEntryUsecase = NewTimeEntryUsecase(u.TimeEntryRepository, u.ProjectUsecase, u.ChangelogRepository)

	syncRepo := postgresql.NewPostgreSQLSyncRepository(test.Database.DB)
	u.SyncUsecase = NewSyncUsecase(syncRepo, u.ChangelogRepository, u.ProjectRepository, u.TimeEntryRepository)
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
