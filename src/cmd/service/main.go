package main

import (
	"net/http"
	"time"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/projects"

	"github.com/gin-gonic/gin"
	ginglog "github.com/szuecs/gin-glog"
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

	projectController := projects.NewController(projects.NewService(databaseService.Database))

	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(gin.Recovery())

	privateGroup := router.Group("/api/v1")
	privateGroup.GET("/private", getPrivate)
	privateGroup.POST("/projects", projectController.AddProject)

	router.GET("/public", getPublic)

	router.Run()
}

func getPublic(context *gin.Context) {
	context.String(http.StatusOK, "Hello world!")
}

func getPrivate(context *gin.Context) {
	context.String(http.StatusOK, "Welcome to the private area. :)")
}
