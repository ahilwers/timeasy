package rest

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/database"
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
	ProjectUsecase   usecase.ProjectUsecase
	TimeEntryUsecase usecase.TimeEntryUsecase
	TeamUsecase      usecase.TeamUsecase
	ProjectHandler   ProjectHandler
	TimeEntryHandler TimeEntryHandler
	TeamHandler      TeamHandler
	Router           *gin.Engine
	tokenVerifier    TokenVerifier
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
	teamRepo := database.NewGormTeamRepository(test.DB)
	t.TeamUsecase = usecase.NewTeamUsecase(teamRepo)

	projectRepo := database.NewGormProjectRepository(test.DB)
	t.ProjectUsecase = usecase.NewProjectUsecase(projectRepo, t.TeamUsecase)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	t.TimeEntryUsecase = usecase.NewTimeEntryUsecase(timeEntryRepo, t.ProjectUsecase)
}

func (t *HandlerTest) initHandlers() {
	authMiddleware := NewJwtAuthMiddleware(t.tokenVerifier)
	t.ProjectHandler = NewProjectHandler(t.tokenVerifier, t.ProjectUsecase)
	t.TimeEntryHandler = NewTimeEntryHandler(t.tokenVerifier, t.TimeEntryUsecase)
	t.TeamHandler = NewTeamHandler(t.tokenVerifier, t.TeamUsecase)

	t.Router = SetupRouter(authMiddleware, t.TeamHandler, t.ProjectHandler, t.TimeEntryHandler)
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
