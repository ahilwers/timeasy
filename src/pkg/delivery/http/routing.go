package http

import (
	"time"

	"github.com/gin-gonic/gin"
	ginglog "github.com/szuecs/gin-glog"
)

func SetupRouter(userHandler UserHandler, projectHandler ProjectHandler, timeEntryHandler TimeEntryHandler) *gin.Engine {
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
	protectedGroup.DELETE("/users/:id", userHandler.DeleteUser)
	protectedGroup.PUT("/users/:id/password", userHandler.UpdatePassword)
	protectedGroup.PUT("/users/:id/roles", userHandler.UpdateRoles)
	protectedGroup.GET("/projects", projectHandler.GetAllProjects)
	protectedGroup.POST("/projects", projectHandler.AddProject)
	protectedGroup.GET("/projects/:id", projectHandler.GetProjectById)
	protectedGroup.PUT("/projects/:id", projectHandler.UpdateProject)
	protectedGroup.DELETE("/projects/:id", projectHandler.DeleteProject)
	protectedGroup.POST("/timeentries", timeEntryHandler.AddTimeEntry)
	protectedGroup.PUT("/timeentries/:id", timeEntryHandler.UpdateTimeEntry)
	protectedGroup.DELETE("/timeentries/:id", timeEntryHandler.DeleteTimeEntry)

	return router
}
