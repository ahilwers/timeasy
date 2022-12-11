package usecase

import (
	"log"
	"os"
	"testing"
	"timeasy-server/pkg/test"
)

func TestMain(m *testing.M) {
	log.Println("Testmain")

	pool, resource := test.SetupDatabase()
	code := m.Run()
	test.TeardownDatabase(pool, resource)

	os.Exit(code)
}
