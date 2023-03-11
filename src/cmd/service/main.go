package main

import (
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

	tokenVerifier := rest.NewJwtTokenVerifier()
	authMiddleware := rest.NewJwtAuthMiddleware(tokenVerifier)

	projectUsecase := usecase.NewProjectUsecase(database.NewGormProjectRepository(databaseService.Database))
	projectHandler := rest.NewProjectHandler(tokenVerifier, projectUsecase)
	userUsecase := usecase.NewUserUsecase(database.NewGormUserRepository(databaseService.Database))
	userHandler := rest.NewUserHandler(tokenVerifier, userUsecase)
	timeEntryUsecase := usecase.NewTimeEntryUsecase(database.NewGormTimeEntryRepository(databaseService.Database), userUsecase, projectUsecase)
	timeEntryHandler := rest.NewTimeEntryHandler(tokenVerifier, timeEntryUsecase)
	teamUsecase := usecase.NewTeamUsecase(database.NewGormTeamRepository(databaseService.Database))
	teamHandler := rest.NewTeamHandler(tokenVerifier, teamUsecase, userUsecase)

	router := rest.SetupRouter(authMiddleware, userHandler, teamHandler, projectHandler, timeEntryHandler)
	router.Run()
}
