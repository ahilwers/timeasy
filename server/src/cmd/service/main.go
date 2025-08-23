package main

import (
	"flag"
	"log"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/database/postgresql"
	"timeasy-server/pkg/external"
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

	providerFactory := external.NewProviderFactory()

	githubClientID := configuration.GitHubClientID
	githubClientSecret := configuration.GitHubClientSecret
	if githubClientID == "" {
		githubClientID = "dummy-client-id"
	}
	if githubClientSecret == "" {
		githubClientSecret = "dummy-client-secret"
	}

	githubProvider := external.NewGitHubProvider(
		githubClientID,
		githubClientSecret,
		configuration.GitHubRedirectURL,
	)
	providerFactory.RegisterProvider(githubProvider)

	gitlabClientID := configuration.GitLabClientID
	gitlabClientSecret := configuration.GitLabClientSecret
	gitlabBaseURL := configuration.GitLabBaseURL
	if gitlabClientID == "" {
		gitlabClientID = "dummy-client-id"
	}
	if gitlabClientSecret == "" {
		gitlabClientSecret = "dummy-client-secret"
	}
	if gitlabBaseURL == "" {
		gitlabBaseURL = "https://gitlab.com"
	}

	gitlabProvider := external.NewGitLabProvider(
		gitlabClientID,
		gitlabClientSecret,
		configuration.GitLabRedirectURL,
		gitlabBaseURL,
	)
	providerFactory.RegisterProvider(gitlabProvider)

	jiraClientID := configuration.JiraClientID
	jiraClientSecret := configuration.JiraClientSecret
	jiraBaseURL := configuration.JiraBaseURL
	if jiraClientID == "" {
		jiraClientID = "dummy-client-id"
	}
	if jiraClientSecret == "" {
		jiraClientSecret = "dummy-client-secret"
	}
	if jiraBaseURL == "" {
		jiraBaseURL = "https://dummy.atlassian.net"
	}

	jiraProvider := external.NewJiraProvider(
		jiraClientID,
		jiraClientSecret,
		configuration.JiraRedirectURL,
		jiraBaseURL,
	)
	providerFactory.RegisterProvider(jiraProvider)

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
	)
	externalIntegrationHandler := rest.NewExternalIntegrationHandler(tokenVerifier, externalIntegrationUsecase)

	router := rest.SetupRouter(authMiddleware, teamHandler, projectHandler, timeEntryHandler, timeEntryExportHandler, syncHandler, weeklyStatisticsHandler, *externalIntegrationHandler, *userExternalAccountHandler)
	router.Run()
}
