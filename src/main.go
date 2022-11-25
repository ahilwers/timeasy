package main

import (
	"net/http"
	"time"
	"timeasy-server/configuration"
	"timeasy-server/database"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	ginglog "github.com/szuecs/gin-glog"
	"github.com/tbaehler/gin-keycloak/pkg/ginkeycloak"
)

func main() {
	configuration, err := configuration.GetConfiguration()
	if err != nil {
		panic(err)
	}

	databaseService := new(database.DatabaseService)
	err = databaseService.Init(configuration.DbHost, configuration.DbName, configuration.DbUser,
		configuration.DbPassword, configuration.DbPort)
	if err != nil {
		panic(err)
	}

	var keycloakconfig = ginkeycloak.KeycloakConfig{
		Url:           configuration.KeyCloakHost,
		Realm:         configuration.KeyCloakRealm,
		FullCertsPath: nil,
	}

	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(ginkeycloak.RequestLogger([]string{"uid"}, "data"))
	router.Use(gin.Recovery())

	privateGroup := router.Group("/api")
	privateGroup.Use(ginkeycloak.Auth(ginkeycloak.AuthCheck(), keycloakconfig))
	privateGroup.GET("/private", getPrivate)

	router.GET("/public", getPublic)

	router.Run()
}

func getPublic(context *gin.Context) {
	context.String(http.StatusOK, "Hello world!")
}

func getPrivate(context *gin.Context) {
	context.String(http.StatusOK, "Welcome to the private area. :)")
	ginToken, _ := context.Get("token")
	token := ginToken.(ginkeycloak.KeyCloakToken)

	glog.Info(token.RealmAccess.Roles)
}
