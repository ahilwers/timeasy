package main

import (
	"flag"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/transport/rest"
	"timeasy-server/pkg/usecase"
)

var databaseService database.DatabaseService

func main() {
	configuration, err := configuration.GetConfiguration()
	if err != nil {
		panic(err)
	}

	err = databaseService.Init(configuration.DbHost, configuration.DbName, configuration.DbUser,
		configuration.DbPassword, configuration.DbPort)
	if err != nil {
		panic(err)
	}

	flag.Parse() // Intialize glog flags

	tokenVerifier := rest.NewKeycloakTokenVerifier(configuration.KeycloakHost, configuration.KeycloakRealm)
	authMiddleware := rest.NewJwtAuthMiddleware(tokenVerifier)

	projectUsecase := usecase.NewProjectUsecase(database.NewGormProjectRepository(databaseService.Database))
	projectHandler := rest.NewProjectHandler(tokenVerifier, projectUsecase)
	timeEntryUsecase := usecase.NewTimeEntryUsecase(database.NewGormTimeEntryRepository(databaseService.Database), projectUsecase)
	timeEntryHandler := rest.NewTimeEntryHandler(tokenVerifier, timeEntryUsecase)
	teamUsecase := usecase.NewTeamUsecase(database.NewGormTeamRepository(databaseService.Database))
	teamHandler := rest.NewTeamHandler(tokenVerifier, teamUsecase)

	router := rest.SetupRouter(authMiddleware, teamHandler, projectHandler, timeEntryHandler)
	router.Run()
}
