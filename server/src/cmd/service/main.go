package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/external"
	"timeasy-server/pkg/sync"
	"timeasy-server/pkg/transport/rest"
	"timeasy-server/pkg/usecase"

	_ "github.com/lib/pq" // PostgreSQL driver
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

	externalConnectionRepository := postgresql.NewPostgreSQLExternalConnectionRepository(databaseService.Database.DB)

	projectHandler := rest.NewProjectHandler(tokenVerifier, projectUsecase, teamUsecase, externalConnectionRepository)

	timeEntryRepository := postgresql.NewPostgreSQLTimeEntryRepository(databaseService.Database.DB)
	timeEntryUsecase := usecase.NewTimeEntryUsecase(timeEntryRepository, projectUsecase, changelogRepository)
	timeEntryHandler := rest.NewTimeEntryHandler(tokenVerifier, timeEntryUsecase)

	syncUsecase := usecase.NewSyncUsecase(postgresql.NewPostgreSQLSyncRepository(databaseService.Database.DB), changelogRepository, projectRepository, timeEntryRepository)
	syncHandler := rest.NewSyncHandler(tokenVerifier, syncUsecase)

	changelogInitUsecase := usecase.NewChangelogInitializationUsecase(changelogRepository, projectRepository, timeEntryRepository, teamRepository)
	if err := changelogInitUsecase.InitializeChangelog(); err != nil {
		log.Printf("Failed to initialize changelog: %v", err)
		panic(err)
	}

	weeklyStatisticsUsecase := usecase.NewWeeklyStatisticsUsecase(timeEntryUsecase)
	weeklyStatisticsHandler := rest.NewWeeklyStatisticsHandler(tokenVerifier, weeklyStatisticsUsecase, projectUsecase)

	timeEntryExportHandler := rest.NewTimeEntryExportHandler(tokenVerifier, timeEntryUsecase)

	// External integration setup
	externalIssueRepository := postgresql.NewPostgreSQLExternalIssueRepository(databaseService.Database.DB)
	userExternalAccountRepository := postgresql.NewPostgreSQLUserExternalAccountRepository(databaseService.Database.DB)

	// Create empty provider factory - providers will be created dynamically as needed
	providerFactory := external.NewProviderFactory()

	userExternalAccountUsecase := usecase.NewUserExternalAccountUseCase(
		userExternalAccountRepository,
		providerFactory,
	)
	userExternalAccountHandler := rest.NewUserExternalAccountHandler(tokenVerifier, userExternalAccountUsecase)

	externalIntegrationUsecase := usecase.NewExternalIntegrationUseCase(
		externalConnectionRepository,
		externalIssueRepository,
		userExternalAccountRepository,
		timeEntryRepository,
		projectRepository,
		providerFactory,
		teamUsecase,
	)
	externalIntegrationHandler := rest.NewExternalIntegrationHandler(tokenVerifier, externalIntegrationUsecase)

	// Setup sync scheduler for automated issue synchronization
	syncConfig := sync.SyncConfig{
		ActiveProjectInterval:   time.Duration(configuration.SyncActiveInterval) * time.Minute,
		RecentProjectInterval:   time.Duration(configuration.SyncRecentInterval) * time.Minute,
		DormantProjectInterval:  time.Duration(configuration.SyncDormantInterval) * time.Minute,
		MaxConcurrentSyncs:      configuration.SyncMaxConcurrent,
		ActivityWindowActive:    4 * time.Hour,
		ActivityWindowRecent:    24 * time.Hour,
	}
	
	syncScheduler := sync.NewSyncScheduler(
		syncConfig,
		externalIntegrationUsecase,
		timeEntryUsecase,
		projectUsecase,
	)
	
	// Set the sync scheduler on the external integration usecase to avoid circular dependency
	externalIntegrationUsecase.SetSyncScheduler(syncScheduler)
	
	// Start the sync scheduler
	syncScheduler.Start()
	defer syncScheduler.Stop()
	
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	router := rest.SetupRouter(authMiddleware, teamHandler, projectHandler, timeEntryHandler, timeEntryExportHandler, syncHandler, weeklyStatisticsHandler, externalIntegrationHandler, userExternalAccountHandler)
	
	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server...")
		router.Run()
	}()
	
	// Wait for interrupt signal
	<-c
	log.Printf("Shutting down gracefully...")
}
