package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"timeasy-server/pkg/test"

	"github.com/gin-gonic/gin"
)

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

type tokenObject struct {
	Token string
}

func Login(router *gin.Engine, username string, password string) (string, error) {
	w := httptest.NewRecorder()

	reader := strings.NewReader(fmt.Sprintf("{\"username\": \"%v\", \"password\": \"%v\"}", username, password))
	loginRequest, err := http.NewRequest("POST", "/api/v1/login", reader)
	if err != nil {
		return "", fmt.Errorf("error creating request for long")
	}
	router.ServeHTTP(w, loginRequest)
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
