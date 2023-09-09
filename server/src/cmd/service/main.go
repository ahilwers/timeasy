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

	teamRepository := database.NewGormTeamRepository(databaseService.Database)
	teamUsecase := usecase.NewTeamUsecase(teamRepository)
	teamHandler := rest.NewTeamHandler(tokenVerifier, teamUsecase)
	projectUsecase := usecase.NewProjectUsecase(database.NewGormProjectRepository(databaseService.Database,
		teamRepository), teamUsecase)
	projectHandler := rest.NewProjectHandler(tokenVerifier, projectUsecase, teamUsecase)
	timeEntryUsecase := usecase.NewTimeEntryUsecase(database.NewGormTimeEntryRepository(databaseService.Database), projectUsecase)
	timeEntryHandler := rest.NewTimeEntryHandler(tokenVerifier, timeEntryUsecase)
	syncHandler := rest.NewSyncHandler(tokenVerifier, timeEntryUsecase)

	router := rest.SetupRouter(authMiddleware, teamHandler, projectHandler, timeEntryHandler, syncHandler)
	router.Run()
}
