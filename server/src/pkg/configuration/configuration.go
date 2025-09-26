package configuration

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/peterbourgon/ff"
)

type Configuration struct {
	DbHost              string
	DbPort              int
	DbName              string
	DbUser              string
	DbPassword          string
	KeycloakHost        string
	KeycloakRealm       string
	SyncActiveInterval  int // Minutes - sync interval for active projects
	SyncRecentInterval  int // Minutes - sync interval for recently active projects
	SyncDormantInterval int // Minutes - sync interval for dormant projects
	SyncMaxConcurrent   int // Maximum concurrent sync operations
	LokiEndpoint        string // Loki server endpoint for logging
	LokiBearerToken     string // Bearer token for Loki authentication
}

func GetConfiguration() (Configuration, error) {
	fs := flag.NewFlagSet("timeasy", flag.ContinueOnError)
	var (
		dbHost              = fs.String("database-host", "localhost", "database host")
		dbPort              = fs.String("database-port", "5432", "database port")
		dbName              = fs.String("database-name", "timeasy", "database name")
		dbUser              = fs.String("database-user", "dbuser", "database user")
		dbPassword          = fs.String("database-password", "dbpassword", "database password")
		keycloakHost        = fs.String("keycloak-host", "http://localhost:8180", "keycloak host")
		keycloakRealm       = fs.String("keycloak-realm", "timeasy", "keycloak realm")
		syncActiveInterval  = fs.String("sync-active-interval", "15", "Sync interval for active projects (minutes)")
		syncRecentInterval  = fs.String("sync-recent-interval", "60", "Sync interval for recently active projects (minutes)")
		syncDormantInterval = fs.String("sync-dormant-interval", "1440", "Sync interval for dormant projects (minutes)")
		syncMaxConcurrent   = fs.String("sync-max-concurrent", "3", "Maximum concurrent sync operations")
		lokiEndpoint        = fs.String("loki-endpoint", "", "Loki server endpoint for logging (e.g., https://loki.example.com/loki/api/v1/push)")
		lokiBearerToken     = fs.String("loki-bearer-token", "", "Bearer token for Loki authentication")

		_ = fs.String("config", "", "config file (optional)")
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

	configuration.LokiEndpoint = *lokiEndpoint
	configuration.LokiBearerToken = *lokiBearerToken

	return configuration, nil
}
