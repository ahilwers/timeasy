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
	return configuration, nil
}
