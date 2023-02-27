package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/test"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var TestUserUsecase usecase.UserUsecase
var TestProjectUsecase usecase.ProjectUsecase
var TestTimeEntryUsecase usecase.TimeEntryUsecase
var TestTeamUsecase usecase.TeamUsecase
var TestUserHandler UserHandler
var TestProjectHandler ProjectHandler
var TestTimeEntryHandler TimeEntryHandler
var TestTeamHandler TeamHandler
var TestRouter *gin.Engine

type ErrorResult struct {
	Error string
}

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}

func SetupTest(tb testing.TB) func(tb testing.TB) {
	tearDown := test.SetupTest(tb)
	initUsecases()
	initHandlers()
	return tearDown
}

func initUsecases() {
	projectRepo := database.NewGormProjectRepository(test.DB)
	TestProjectUsecase = usecase.NewProjectUsecase(projectRepo)

	userRepo := database.NewGormUserRepository(test.DB)
	TestUserUsecase = usecase.NewUserUsecase(userRepo)

	timeEntryRepo := database.NewGormTimeEntryRepository(test.DB)
	TestTimeEntryUsecase = usecase.NewTimeEntryUsecase(timeEntryRepo, TestUserUsecase, TestProjectUsecase)

	teamRepo := database.NewGormTeamRepository(test.DB)
	TestTeamUsecase = usecase.NewTeamUsecase(teamRepo)
}

func initHandlers() {
	TestUserHandler = NewUserHandler(TestUserUsecase)
	TestProjectHandler = NewProjectHandler(TestProjectUsecase)
	TestTimeEntryHandler = NewTimeEntryHandler(TestTimeEntryUsecase)
	TestTeamHandler = NewTeamHandler(TestTeamUsecase, TestUserUsecase)

	TestRouter = SetupRouter(TestUserHandler, TestTeamHandler, TestProjectHandler, TestTimeEntryHandler)
}

type tokenObject struct {
	Token string
}

func Login(username string, password string) (string, error) {
	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"username\": \"%v\", \"password\": \"%v\"}", username, password))
	loginRequest, err := http.NewRequest("POST", "/api/v1/login", reader)
	if err != nil {
		return "", fmt.Errorf("error creating request for long")
	}
	TestRouter.ServeHTTP(w, loginRequest)
	if w.Code != 200 {
		return "", fmt.Errorf("error logging in: %v", w.Code)
	}
	var tokenObject tokenObject
	json.Unmarshal(w.Body.Bytes(), &tokenObject)
	return tokenObject.Token, nil
}

func AddToken(req *http.Request, token string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
