package configuration

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/peterbourgon/ff"
)

type Configuration struct {
	DbHost        string
	DbPort        int
	DbName        string
	DbUser        string
	DbPassword    string
	KeycloakHost  string
	KeycloakRealm string
	
	// External provider OAuth settings
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string
	
	GitLabClientID     string
	GitLabClientSecret string
	GitLabRedirectURL  string
	GitLabBaseURL      string
	
	JiraClientID      string
	JiraClientSecret  string
	JiraRedirectURL   string
	JiraBaseURL       string
	
	// Sync scheduler settings
	SyncActiveInterval   int // Minutes - sync interval for active projects
	SyncRecentInterval   int // Minutes - sync interval for recently active projects  
	SyncDormantInterval  int // Minutes - sync interval for dormant projects
	SyncMaxConcurrent    int // Maximum concurrent sync operations
}

func GetConfiguration() (Configuration, error) {
	fs := flag.NewFlagSet("timeasy", flag.ContinueOnError)
	var (
		dbHost        = fs.String("database-host", "localhost", "database host")
		dbPort        = fs.String("database-port", "5432", "database port")
		dbName        = fs.String("database-name", "timeasy", "database name")
		dbUser        = fs.String("database-user", "dbuser", "database user")
		dbPassword    = fs.String("database-password", "dbpassword", "database password")
		keycloakHost  = fs.String("keycloak-host", "http://localhost:8180", "keycloak host")
		keycloakRealm = fs.String("keycloak-realm", "timeasy", "keycloak realm")
		
		// GitHub OAuth settings
		githubClientID     = fs.String("github-client-id", "", "GitHub OAuth client ID")
		githubClientSecret = fs.String("github-client-secret", "", "GitHub OAuth client secret")
		githubRedirectURL  = fs.String("github-redirect-url", "", "GitHub OAuth redirect URL")
		
		// GitLab OAuth settings
		gitlabClientID     = fs.String("gitlab-client-id", "", "GitLab OAuth client ID")
		gitlabClientSecret = fs.String("gitlab-client-secret", "", "GitLab OAuth client secret")
		gitlabRedirectURL  = fs.String("gitlab-redirect-url", "", "GitLab OAuth redirect URL")
		gitlabBaseURL      = fs.String("gitlab-base-url", "https://gitlab.com", "GitLab base URL")
		
		// Jira OAuth settings
		jiraClientID      = fs.String("jira-client-id", "", "Jira OAuth client ID")
		jiraClientSecret  = fs.String("jira-client-secret", "", "Jira OAuth client secret")
		jiraRedirectURL   = fs.String("jira-redirect-url", "", "Jira OAuth redirect URL")
		jiraBaseURL       = fs.String("jira-base-url", "", "Jira base URL")
		
		// Sync scheduler settings  
		syncActiveInterval   = fs.String("sync-active-interval", "15", "Sync interval for active projects (minutes)")
		syncRecentInterval   = fs.String("sync-recent-interval", "60", "Sync interval for recently active projects (minutes)")
		syncDormantInterval  = fs.String("sync-dormant-interval", "1440", "Sync interval for dormant projects (minutes)")
		syncMaxConcurrent    = fs.String("sync-max-concurrent", "3", "Maximum concurrent sync operations")
		
		_             = fs.String("config", "", "config file (optional)")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("TIMEASY"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)

	var configuration Configuration
	configuration.DbName = *dbName
	configuration.DbUser = *dbUser
	configuration.DbPassword = *dbPassword
	configuration.DbHost = *dbHost
	port, err := strconv.Atoi(*dbPort)
	if err != nil {
		return configuration, fmt.Errorf("the specified port is invalid: %w", err)
	}
	configuration.DbPort = port
	configuration.KeycloakHost = *keycloakHost
	configuration.KeycloakRealm = *keycloakRealm
	
	// External provider settings
	configuration.GitHubClientID = *githubClientID
	configuration.GitHubClientSecret = *githubClientSecret
	configuration.GitHubRedirectURL = *githubRedirectURL
	
	configuration.GitLabClientID = *gitlabClientID
	configuration.GitLabClientSecret = *gitlabClientSecret
	configuration.GitLabRedirectURL = *gitlabRedirectURL
	configuration.GitLabBaseURL = *gitlabBaseURL
	
	configuration.JiraClientID = *jiraClientID
	configuration.JiraClientSecret = *jiraClientSecret
	configuration.JiraRedirectURL = *jiraRedirectURL
	configuration.JiraBaseURL = *jiraBaseURL
	
	// Parse sync scheduler settings
	syncActive, err := strconv.Atoi(*syncActiveInterval)
	if err != nil {
		return configuration, fmt.Errorf("invalid sync active interval: %w", err)
	}
	configuration.SyncActiveInterval = syncActive
	
	syncRecent, err := strconv.Atoi(*syncRecentInterval)
	if err != nil {
		return configuration, fmt.Errorf("invalid sync recent interval: %w", err)
	}
	configuration.SyncRecentInterval = syncRecent
	
	syncDormant, err := strconv.Atoi(*syncDormantInterval)
	if err != nil {
		return configuration, fmt.Errorf("invalid sync dormant interval: %w", err)
	}
	configuration.SyncDormantInterval = syncDormant
	
	syncConcurrent, err := strconv.Atoi(*syncMaxConcurrent)
	if err != nil {
		return configuration, fmt.Errorf("invalid sync max concurrent: %w", err)
	}
	configuration.SyncMaxConcurrent = syncConcurrent
	
	return configuration, nil
}
