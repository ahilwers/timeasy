package configuration

import (
	"flag"
	"fmt"
	"github.com/peterbourgon/ff"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Configuration struct {
	ServerPort                  int // HTTP server port
	DbHost                      string
	DbPort                      int
	DbName                      string
	DbUser                      string
	DbPassword                  string
	KeycloakHost                string
	KeycloakRealm               string
	SyncActiveInterval          int      // Minutes - sync interval for active projects
	SyncRecentInterval          int      // Minutes - sync interval for recently active projects
	SyncDormantInterval         int      // Minutes - sync interval for dormant projects
	SyncMaxConcurrent           int      // Maximum concurrent sync operations
	LokiEndpoint                string   // Loki server endpoint for logging
	LokiBearerToken             string   // Bearer token for Loki authentication
	LogLevel                    string   // Log level: DEBUG, INFO, WARN, ERROR
	RevenueCatAPIKey            string   // RevenueCat v2 secret API key (empty = billing disabled, self-hosted mode)
	RevenueCatWebAPIKey         string   // Public RevenueCat Web Billing SDK key, served to the web client
	RevenueCatProjectID         string   // RevenueCat project ID for v2 customer lookup
	RevenueCatEntitlementID     string   // RevenueCat entitlement ID (entl...) that grants access
	SubscriptionCacheTTLMinutes int      // Minutes a subscription check result is cached
	SubscriptionAllowlist       []string // Keycloak user IDs with full access without a subscription (owner/testers)
}

func GetConfiguration() (Configuration, error) {
	fs := flag.NewFlagSet("timeasy", flag.ContinueOnError)
	var (
		serverPort              = fs.String("server-port", "8080", "HTTP server port")
		dbHost                  = fs.String("database-host", "localhost", "database host")
		dbPort                  = fs.String("database-port", "5432", "database port")
		dbName                  = fs.String("database-name", "timeasy", "database name")
		dbUser                  = fs.String("database-user", "dbuser", "database user")
		dbPassword              = fs.String("database-password", "dbpassword", "database password")
		keycloakHost            = fs.String("keycloak-host", "http://localhost:8180", "keycloak host")
		keycloakRealm           = fs.String("keycloak-realm", "timeasy", "keycloak realm")
		syncActiveInterval      = fs.String("sync-active-interval", "15", "Sync interval for active projects (minutes)")
		syncRecentInterval      = fs.String("sync-recent-interval", "60", "Sync interval for recently active projects (minutes)")
		syncDormantInterval     = fs.String("sync-dormant-interval", "1440", "Sync interval for dormant projects (minutes)")
		syncMaxConcurrent       = fs.String("sync-max-concurrent", "3", "Maximum concurrent sync operations")
		lokiEndpoint            = fs.String("loki-endpoint", "", "Loki server endpoint for logging (e.g., https://loki.example.com/loki/api/v1/push)")
		lokiBearerToken         = fs.String("loki-bearer-token", "", "Bearer token for Loki authentication")
		logLevel                = fs.String("log-level", "INFO", "Log level: DEBUG, INFO, WARN, ERROR")
		revenueCatAPIKey        = fs.String("revenuecat-api-key", "", "RevenueCat v2 secret API key (empty = billing disabled, self-hosted mode)")
		revenueCatWebAPIKey     = fs.String("revenuecat-web-api-key", "", "Public RevenueCat Web Billing SDK key served to the web client")
		revenueCatProjectID     = fs.String("revenuecat-project-id", "", "RevenueCat project ID required when billing is enabled")
		revenueCatEntitlementID = fs.String("revenuecat-entitlement-id", "", "RevenueCat entitlement ID (entl...) required when billing is enabled")
		subscriptionCacheTTL    = fs.String("subscription-cache-ttl-minutes", "15", "Minutes a subscription check result is cached")
		subscriptionAllowlist   = fs.String("subscription-allowlist", "", "Comma-separated Keycloak user IDs that get full access without a subscription (owner/testers)")

		_ = fs.String("config", "", "config file (optional)")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("TIMEASY"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)

	var configuration Configuration

	// Parse server port
	srvPort, err := strconv.Atoi(*serverPort)
	if err != nil {
		return configuration, fmt.Errorf("the specified server port is invalid: %w", err)
	}
	configuration.ServerPort = srvPort

	configuration.DbName = *dbName
	configuration.DbUser = *dbUser
	configuration.DbPassword = *dbPassword
	configuration.DbHost = *dbHost
	port, err := strconv.Atoi(*dbPort)
	if err != nil {
		return configuration, fmt.Errorf("the specified database port is invalid: %w", err)
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
	configuration.LogLevel = *logLevel
	configuration.RevenueCatAPIKey = *revenueCatAPIKey
	configuration.RevenueCatWebAPIKey = strings.TrimSpace(*revenueCatWebAPIKey)
	configuration.RevenueCatProjectID = strings.TrimSpace(*revenueCatProjectID)
	configuration.RevenueCatEntitlementID = strings.TrimSpace(*revenueCatEntitlementID)
	if err := validateRevenueCatConfiguration(configuration); err != nil {
		return configuration, err
	}

	subscriptionCacheTTLMinutes, err := strconv.Atoi(*subscriptionCacheTTL)
	if err != nil {
		return configuration, fmt.Errorf("invalid subscription cache TTL: %w", err)
	}
	configuration.SubscriptionCacheTTLMinutes = subscriptionCacheTTLMinutes
	configuration.SubscriptionAllowlist = parseAllowlist(*subscriptionAllowlist)

	return configuration, nil
}

// validateRevenueCatConfiguration ensures all identifiers needed for v2
// customer lookups are provided when billing is enabled. With no API key,
// RevenueCat configuration is optional for self-hosted instances.
func validateRevenueCatConfiguration(configuration Configuration) error {
	if configuration.RevenueCatAPIKey == "" {
		return nil
	}
	if configuration.RevenueCatProjectID == "" {
		return fmt.Errorf("revenuecat-project-id is required when revenuecat-api-key is set")
	}
	if configuration.RevenueCatEntitlementID == "" {
		return fmt.Errorf("revenuecat-entitlement-id is required when revenuecat-api-key is set")
	}
	return nil
}

// parseAllowlist splits a comma-separated list of user IDs, trimming
// whitespace and dropping empty entries.
func parseAllowlist(raw string) []string {
	result := make([]string, 0, 4)
	for _, part := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (c *Configuration) ParseLogLevel() slog.Level {
	switch strings.ToUpper(c.LogLevel) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
