package http

import (
	"time"

	"github.com/gin-gonic/gin"
	ginglog "github.com/szuecs/gin-glog"
)

func SetupRouter(userHandler UserHandler, projectHandler ProjectHandler) *gin.Engine {
	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(gin.Recovery())

	publicGroup := router.Group("/api/v1")
	publicGroup.POST("/signup", userHandler.Signup)
	publicGroup.POST("/login", userHandler.Login)

	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(JwtAuthMiddleware())
	protectedGroup.GET("/users/:id", userHandler.GetUserById)
	protectedGroup.GET("/users", userHandler.GetAllUsers)
	protectedGroup.PUT("/users/:id", userHandler.UpdateUser)
	protectedGroup.PUT("/users/:id/password", userHandler.UpdatePassword)
	protectedGroup.POST("/projects", projectHandler.AddProject)

	return router
}
