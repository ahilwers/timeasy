package rest

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/external"
	"timeasy-server/pkg/test"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}

type tokenVerifierMock struct {
	mock.Mock
}

func (m *tokenVerifierMock) VerifyToken(c *gin.Context) (AuthToken, error) {
	args := m.Called(c)
	return args.Get(0).(AuthToken), args.Error(1)
}

type authTokenMock struct {
	mock.Mock
}

func (t *authTokenMock) GetUserId() (uuid.UUID, error) {
	args := t.Called()
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (t *authTokenMock) GetRoles() ([]string, error) {
	args := t.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (t *authTokenMock) HasRole(role string) (bool, error) {
	args := t.Called(role)
	return args.Get(0).(bool), args.Error(1)
}

type HandlerTest struct {
	ProjectUsecase          usecase.ProjectUsecase
	TimeEntryUsecase        usecase.TimeEntryUsecase
	TeamUsecase             usecase.TeamUsecase
	SyncUsecase             usecase.SyncUsecase
	WeeklyStatisticsUsecase *usecase.WeeklyStatisticsUsecase
	ProjectHandler          ProjectHandler
	TimeEntryHandler        TimeEntryHandler
	TeamHandler             TeamHandler
	SyncHandler             SyncHandler
	WeeklyStatisticsHandler WeeklyStatisticsHandler
	TimeEntryExportHandler  TimeEntryExportHandler
	Router                  *gin.Engine
	tokenVerifier           TokenVerifier
}

type ErrorResult struct {
	Error string
}

func NewHandlerTest(tokenVerifier TokenVerifier) *HandlerTest {
	return &HandlerTest{
		tokenVerifier: tokenVerifier,
	}
}

func (t *HandlerTest) SetupTest(tb testing.TB) func(tb testing.TB) {
	tearDown := test.SetupTest(tb)
	t.initUsecases()
	t.initHandlers()
	return tearDown
}

func (t *HandlerTest) initUsecases() {
	teamRepo := postgresql.NewPostgreSQLTeamRepository(test.Database.DB)
	changelogRepo := postgresql.NewPostgreSQLChangelogRepository(test.Database.DB)
	t.TeamUsecase = usecase.NewTeamUsecase(teamRepo, changelogRepo)

	projectRepo := postgresql.NewPostgreSQLProjectRepository(test.Database.DB, teamRepo)
	t.ProjectUsecase = usecase.NewProjectUsecase(projectRepo, t.TeamUsecase, changelogRepo)

	timeEntryRepo := postgresql.NewPostgreSQLTimeEntryRepository(test.Database.DB)
	t.TimeEntryUsecase = usecase.NewTimeEntryUsecase(timeEntryRepo, t.ProjectUsecase, changelogRepo)

	syncRepo := postgresql.NewPostgreSQLSyncRepository(test.Database.DB)
	t.SyncUsecase = usecase.NewSyncUsecase(syncRepo, changelogRepo, projectRepo, timeEntryRepo)

	t.WeeklyStatisticsUsecase = usecase.NewWeeklyStatisticsUsecase(t.TimeEntryUsecase)
}

func (t *HandlerTest) initHandlers() {
	authMiddleware := NewJwtAuthMiddleware(t.tokenVerifier)
	
	// Create external connection repository for project handler
	externalConnectionRepo := postgresql.NewPostgreSQLExternalConnectionRepository(test.Database.DB)
	t.ProjectHandler = NewProjectHandler(t.tokenVerifier, t.ProjectUsecase, t.TeamUsecase, externalConnectionRepo)
	
	t.TimeEntryHandler = NewTimeEntryHandler(t.tokenVerifier, t.TimeEntryUsecase)
	t.TeamHandler = NewTeamHandler(t.tokenVerifier, t.TeamUsecase)
	t.SyncHandler = NewSyncHandler(t.tokenVerifier, t.SyncUsecase)
	t.WeeklyStatisticsHandler = NewWeeklyStatisticsHandler(t.tokenVerifier, t.WeeklyStatisticsUsecase, t.ProjectUsecase)
	t.TimeEntryExportHandler = NewTimeEntryExportHandler(t.tokenVerifier, t.TimeEntryUsecase)
	
	// Create external integration handlers
	externalIssueRepo := postgresql.NewPostgreSQLExternalIssueRepository(test.Database.DB)
	userExternalAccountRepo := postgresql.NewPostgreSQLUserExternalAccountRepository(test.Database.DB)
	timeEntryRepo := postgresql.NewPostgreSQLTimeEntryRepository(test.Database.DB)
	projectRepo := postgresql.NewPostgreSQLProjectRepository(test.Database.DB, postgresql.NewPostgreSQLTeamRepository(test.Database.DB))
	
	// Create provider factory (can be nil for tests)
	providerFactory := &external.ProviderFactory{}
	
	externalIntegrationUsecase := usecase.NewExternalIntegrationUseCase(
		externalConnectionRepo, 
		externalIssueRepo, 
		userExternalAccountRepo,
		timeEntryRepo,
		projectRepo,
		providerFactory)
	userExternalAccountUsecase := usecase.NewUserExternalAccountUseCase(userExternalAccountRepo, providerFactory)
	
	externalIntegrationHandler := NewExternalIntegrationHandler(t.tokenVerifier, externalIntegrationUsecase)
	userExternalAccountHandler := NewUserExternalAccountHandler(t.tokenVerifier, userExternalAccountUsecase)

	t.Router = SetupRouter(authMiddleware, t.TeamHandler, t.ProjectHandler, t.TimeEntryHandler, t.TimeEntryExportHandler, t.SyncHandler, t.WeeklyStatisticsHandler, externalIntegrationHandler, userExternalAccountHandler)
}

func AssertErrorMessageEquals(t *testing.T, responseBody []byte, expectedMessage string) {
	actualMessage := GetErrorMessageFromResponse(t, responseBody)
	assert.Equal(t, expectedMessage, actualMessage)
}

func GetErrorMessageFromResponse(t *testing.T, responseBody []byte) string {
	var errorResult ErrorResult
	err := json.Unmarshal(responseBody, &errorResult)
	assert.Nil(t, err)
	return errorResult.Error
}
