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
	userUsecase := usecase.NewUserUsecase(database.NewGormUserRepository(databaseService.Database))
	userHandler := http.NewUserHandler(userUsecase)

	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(gin.Recovery())

	publicGroup := router.Group("/api/v1")
	publicGroup.POST("/signup", userHandler.Signup)
	publicGroup.POST("/login", userHandler.Login)

	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(http.JwtAuthMiddleware())
	protectedGroup.POST("/projects", projectHandler.AddProject)

	router.Run()
}
