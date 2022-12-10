package main

import (
	"time"
	"timeasy-server/pkg/configuration"
	"timeasy-server/pkg/database"
	"timeasy-server/pkg/delivery/http"
	"timeasy-server/pkg/usecase"

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

	projectUsecase := usecase.NewProjectUsecase(database.NewGormProjectRepository(databaseService.Database))
	projectHandler := http.NewProjectHandler(projectUsecase)

	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(gin.Recovery())

	privateGroup := router.Group("/api/v1")
	privateGroup.POST("/projects", projectHandler.AddProject)

	router.Run()
}
