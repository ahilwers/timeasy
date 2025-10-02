package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/external"
	"timeasy-server/pkg/logging"
	"timeasy-server/pkg/sync"
	"timeasy-server/pkg/transport/rest"
	"timeasy-server/pkg/usecase"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var databaseService database.DatabaseService

func main() {
	configuration, err := configuration.GetConfiguration()
	if err != nil {
		slog.Error("Failed to get configuration", "error", err)
		panic(err)
	}

	// Setup base structured logging with colorful console output using configured log level
	logLevel := configuration.ParseLogLevel()
	colorfulHandler := rest.NewColorfulHandler(os.Stdout, &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: false, // Disable source info for cleaner console output
	})
	slog.SetDefault(slog.New(colorfulHandler)) // will be overriden by Loki logger later on

	// Setup Loki logging with fallback to colorful console handler
	lokiLogger, err := logging.NewLokiLogger(configuration, colorfulHandler)
	if err != nil {
		slog.Error("Failed to initialize Loki logger", "error", err)
		panic(err)
	}

	// Set up the combined logger as the default
	logger := slog.New(lokiLogger.Handler())
	slog.SetDefault(logger)

	// Ensure graceful shutdown of Loki logger
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := lokiLogger.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown Loki logger", "error", err)
		}
	}()

	slog.Debug("Debug logging enabled")

	slog.Info("Connecting to database",
		"host", configuration.DbHost,
		"port", configuration.DbPort,
		"database", configuration.DbName)
	err = databaseService.Init(configuration.DbHost, configuration.DbName, configuration.DbUser,
		configuration.DbPassword, configuration.DbPort)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}

	slog.Info("Authentication server configured", "keycloak_host", configuration.KeycloakHost)

	flag.Parse()

	tokenVerifier := rest.NewKeycloakTokenVerifier(configuration.KeycloakHost, configuration.KeycloakRealm)
	authMiddleware := rest.NewJwtAuthMiddleware(tokenVerifier)

	changelogRepository := postgresql.NewPostgreSQLChangelogRepository(databaseService.Database.DB)

	teamRepository := postgresql.NewPostgreSQLTeamRepository(databaseService.Database.DB)
	teamUsecase := usecase.NewTeamUsecase(teamRepository, changelogRepository)
	teamHandler := rest.NewTeamHandler(tokenVerifier, teamUsecase)

	projectRepository := postgresql.NewPostgreSQLProjectRepository(databaseService.Database.DB, teamRepository)
	projectUsecase := usecase.NewProjectUsecase(projectRepository, teamUsecase, changelogRepository)

	externalConnectionRepository := postgresql.NewPostgreSQLExternalConnectionRepository(databaseService.Database.DB)

	externalIssueRepository := postgresql.NewPostgreSQLExternalIssueRepository(databaseService.Database.DB)
	userExternalAccountRepository := postgresql.NewPostgreSQLUserExternalAccountRepository(databaseService.Database.DB)

	// Create empty provider factory - providers will be created dynamically as needed
	providerFactory := external.NewProviderFactory()

	userExternalAccountUsecase := usecase.NewUserExternalAccountUseCase(
		userExternalAccountRepository,
		providerFactory,
	)
	userExternalAccountHandler := rest.NewUserExternalAccountHandler(tokenVerifier, userExternalAccountUsecase)

	timeEntryRepository := postgresql.NewPostgreSQLTimeEntryRepository(databaseService.Database.DB)

	externalIntegrationUsecase := usecase.NewExternalIntegrationUseCase(
		externalConnectionRepository,
		externalIssueRepository,
		userExternalAccountRepository,
		timeEntryRepository,
		projectRepository,
		providerFactory,
		teamUsecase,
	)

	projectHandler := rest.NewProjectHandler(tokenVerifier, projectUsecase, teamUsecase, externalConnectionRepository)

	timeEntryUsecase := usecase.NewTimeEntryUsecase(timeEntryRepository, projectUsecase, changelogRepository, externalIntegrationUsecase)
	timeEntryHandler := rest.NewTimeEntryHandler(tokenVerifier, timeEntryUsecase)

	syncUsecase := usecase.NewSyncUsecase(postgresql.NewPostgreSQLSyncRepository(databaseService.Database.DB), changelogRepository, projectRepository, timeEntryRepository, externalIntegrationUsecase)
	syncHandler := rest.NewSyncHandler(tokenVerifier, syncUsecase)

	changelogInitUsecase := usecase.NewChangelogInitializationUsecase(changelogRepository, projectRepository, timeEntryRepository, teamRepository)
	if err := changelogInitUsecase.InitializeChangelog(); err != nil {
		slog.Error("Failed to initialize changelog", "error", err)
		panic(err)
	}

	weeklyStatisticsUsecase := usecase.NewWeeklyStatisticsUsecase(timeEntryUsecase)
	weeklyStatisticsHandler := rest.NewWeeklyStatisticsHandler(tokenVerifier, weeklyStatisticsUsecase, projectUsecase)

	timeEntryExportHandler := rest.NewTimeEntryExportHandler(tokenVerifier, timeEntryUsecase)
	externalIntegrationHandler := rest.NewExternalIntegrationHandler(tokenVerifier, externalIntegrationUsecase)

	// Setup sync scheduler for automated issue synchronization
	syncConfig := sync.SyncConfig{
		ActiveProjectInterval:  time.Duration(configuration.SyncActiveInterval) * time.Minute,
		RecentProjectInterval:  time.Duration(configuration.SyncRecentInterval) * time.Minute,
		DormantProjectInterval: time.Duration(configuration.SyncDormantInterval) * time.Minute,
		MaxConcurrentSyncs:     configuration.SyncMaxConcurrent,
		ActivityWindowActive:   4 * time.Hour,
		ActivityWindowRecent:   24 * time.Hour,
	}

	syncScheduler := sync.NewSyncScheduler(
		syncConfig,
		externalIntegrationUsecase,
		timeEntryUsecase,
		projectUsecase,
	)

	externalIntegrationUsecase.SetSyncScheduler(syncScheduler)

	// Start the sync scheduler
	syncScheduler.Start()
	defer syncScheduler.Stop()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	router := rest.SetupRouter(authMiddleware, logger, teamHandler, projectHandler, timeEntryHandler, timeEntryExportHandler, syncHandler, weeklyStatisticsHandler, externalIntegrationHandler, userExternalAccountHandler)

	go func() {
		slog.Info("Starting HTTP server", "port", "8080")
		router.Run()
	}()

	// Wait for interrupt signal
	<-c
	slog.Info("Shutting down gracefully...")
}
