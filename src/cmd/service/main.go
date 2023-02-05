package main

import (
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/delivery/http"
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

	projectUsecase := usecase.NewProjectUsecase(database.NewGormProjectRepository(databaseService.Database))
	projectHandler := http.NewProjectHandler(projectUsecase)
	userUsecase := usecase.NewUserUsecase(database.NewGormUserRepository(databaseService.Database))
	userHandler := http.NewUserHandler(userUsecase)
	timeEntryUsecase := usecase.NewTimeEntryUsecase(database.NewGormTimeEntryRepository(databaseService.Database), userUsecase, projectUsecase)
	timeEntryHandler := http.NewTimeEntryHandler(timeEntryUsecase)
	teamUsecase := usecase.NewTeamUsecase(database.NewGormTeamRepository(databaseService.Database))
	teamHandler := http.NewTeamHandler(teamUsecase)

	router := http.SetupRouter(userHandler, teamHandler, projectHandler, timeEntryHandler)
	router.Run()
}
