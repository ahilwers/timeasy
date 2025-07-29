package main

import (
	"flag"
	"log"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/transport/rest"
	"timeasy-server/pkg/usecase"
)

var databaseService database.DatabaseService

func main() {
	configuration, err := configuration.GetConfiguration()
	if err != nil {
		panic(err)
	}

	log.Printf("Connecting to database at %v:%v\n", configuration.DbHost, configuration.DbPort)
	err = databaseService.Init(configuration.DbHost, configuration.DbName, configuration.DbUser,
		configuration.DbPassword, configuration.DbPort)
	if err != nil {
		panic(err)
	}

	log.Printf("Authentication server is at %v\n", configuration.KeycloakHost)

	flag.Parse() // Intialize glog flags

	tokenVerifier := rest.NewKeycloakTokenVerifier(configuration.KeycloakHost, configuration.KeycloakRealm)
	authMiddleware := rest.NewJwtAuthMiddleware(tokenVerifier)

	changelogRepository := postgresql.NewPostgreSQLChangelogRepository(databaseService.Database.DB)

	teamRepository := postgresql.NewPostgreSQLTeamRepository(databaseService.Database.DB)
	teamUsecase := usecase.NewTeamUsecase(teamRepository, changelogRepository)
	teamHandler := rest.NewTeamHandler(tokenVerifier, teamUsecase)

	projectRepository := postgresql.NewPostgreSQLProjectRepository(databaseService.Database.DB, teamRepository)
	projectUsecase := usecase.NewProjectUsecase(projectRepository, teamUsecase, changelogRepository)
	projectHandler := rest.NewProjectHandler(tokenVerifier, projectUsecase, teamUsecase)

	timeEntryRepository := postgresql.NewPostgreSQLTimeEntryRepository(databaseService.Database.DB)
	timeEntryUsecase := usecase.NewTimeEntryUsecase(timeEntryRepository, projectUsecase, changelogRepository)
	timeEntryHandler := rest.NewTimeEntryHandler(tokenVerifier, timeEntryUsecase)

	syncUsecase := usecase.NewSyncUsecase(postgresql.NewPostgreSQLSyncRepository(databaseService.Database.DB), changelogRepository, projectRepository, timeEntryRepository)
	syncHandler := rest.NewSyncHandler(tokenVerifier, syncUsecase)

	weeklyStatisticsUsecase := usecase.NewWeeklyStatisticsUsecase(timeEntryUsecase)
	weeklyStatisticsHandler := rest.NewWeeklyStatisticsHandler(tokenVerifier, weeklyStatisticsUsecase, projectUsecase)

	timeEntryExportHandler := rest.NewTimeEntryExportHandler(tokenVerifier, timeEntryUsecase)

	router := rest.SetupRouter(authMiddleware, teamHandler, projectHandler, timeEntryHandler, timeEntryExportHandler, syncHandler, weeklyStatisticsHandler)
	router.Run()
}
