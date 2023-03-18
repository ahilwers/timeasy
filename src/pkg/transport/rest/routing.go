package rest

import (
	"time"

	"github.com/gin-gonic/gin"
	ginglog "github.com/szuecs/gin-glog"
)

func SetupRouter(authMiddleware AuthMiddleware, teamHandler TeamHandler, projectHandler ProjectHandler, timeEntryHandler TimeEntryHandler) *gin.Engine {
	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(gin.Recovery())

	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(authMiddleware.HandlerFunc())
	protectedGroup.GET("/projects", projectHandler.GetAllProjects)
	protectedGroup.POST("/projects", projectHandler.AddProject)
	protectedGroup.GET("/projects/:id", projectHandler.GetProjectById)
	protectedGroup.PUT("/projects/:id", projectHandler.UpdateProject)
	protectedGroup.DELETE("/projects/:id", projectHandler.DeleteProject)
	protectedGroup.GET("/timeentries/:id", timeEntryHandler.GetTimeEntryById)
	protectedGroup.GET("/timeentries", timeEntryHandler.GetAllTimeEntries)
	protectedGroup.POST("/timeentries", timeEntryHandler.AddTimeEntry)
	protectedGroup.PUT("/timeentries/:id", timeEntryHandler.UpdateTimeEntry)
	protectedGroup.DELETE("/timeentries/:id", timeEntryHandler.DeleteTimeEntry)
	protectedGroup.GET("/teams/:id", teamHandler.GetTeamById)
	protectedGroup.GET("/teams", teamHandler.GetAllTeams)
	protectedGroup.POST("/teams", teamHandler.AddTeam)
	protectedGroup.PUT("/teams/:id", teamHandler.UpdateTeam)
	protectedGroup.DELETE("/teams/:id", teamHandler.DeleteTeam)
	protectedGroup.POST("/teams/:id/users", teamHandler.AddUserToTeam)
	protectedGroup.DELETE("/teams/:id/users/:userId", teamHandler.DeleteUserFromTeam)
	protectedGroup.PUT("/teams/:id/users/:userId/roles", teamHandler.UpdateUserRolesInTeam)

	return router
}
