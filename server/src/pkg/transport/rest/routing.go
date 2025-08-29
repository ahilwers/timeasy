package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ginglog "github.com/szuecs/gin-glog"
)

func SetupRouter(authMiddleware AuthMiddleware, teamHandler TeamHandler, projectHandler ProjectHandler, timeEntryHandler TimeEntryHandler, timeEntryExportHandler TimeEntryExportHandler, syncHandler SyncHandler, weeklyStatisticsHandler WeeklyStatisticsHandler, externalIntegrationHandler ExternalIntegrationHandler, userExternalAccountHandler UserExternalAccountHandler) *gin.Engine {
	router := gin.Default()

	router.Use(ginglog.Logger(3 * time.Second))
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(authMiddleware.HandlerFunc())
	protectedGroup.GET("/projects", projectHandler.GetAllProjects)
	protectedGroup.POST("/projects", projectHandler.AddProject)
	protectedGroup.GET("/projects/:id", projectHandler.GetProjectById)
	protectedGroup.PUT("/projects/:id", projectHandler.UpdateProject)
	protectedGroup.POST("/projects/team", projectHandler.AssignProjectToTeam)
	protectedGroup.DELETE("/projects/:id", projectHandler.DeleteProject)
	protectedGroup.GET("/timeentries/:id", timeEntryHandler.GetTimeEntryById)
	protectedGroup.GET("/timeentries/lastopen", timeEntryHandler.GetLastOpenTimeEntry)
	protectedGroup.GET("/timeentries", timeEntryHandler.GetAllTimeEntries)
	protectedGroup.GET("/timeentries/ascsv", timeEntryExportHandler.ExportTimeEntriesToCsv)
	protectedGroup.GET("/timeentries/asxlsx", timeEntryExportHandler.ExportTimeEntriesToXls)
	protectedGroup.GET("/timeentries/asxlsxonelineperday", timeEntryExportHandler.ExportTimeEntriesToXlsOneLinePerDay)
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
	protectedGroup.GET("/sync/changed/:sinceChangeLogEntry", syncHandler.GetChangedEntries)
	protectedGroup.POST("/sync/changed", syncHandler.SendLocallyChangedEntries)
	protectedGroup.GET("/weeklystatistics/:week/:year", weeklyStatisticsHandler.GetWeeklyStatistics)
	protectedGroup.GET("/currentweeknumber", weeklyStatisticsHandler.GetCurrentWeekNumber)

	// User external accounts endpoints
	protectedGroup.POST("/user/external-accounts", userExternalAccountHandler.CreateAccount)
	protectedGroup.GET("/user/external-accounts", userExternalAccountHandler.GetAccounts)
	protectedGroup.GET("/user/external-accounts/provider/:provider", userExternalAccountHandler.GetAccountsByProvider)
	protectedGroup.PUT("/user/external-accounts/:accountId", userExternalAccountHandler.UpdateAccount)
	protectedGroup.DELETE("/user/external-accounts/:accountId", userExternalAccountHandler.DeleteAccount)
	protectedGroup.POST("/user/external-accounts/:accountId/test", userExternalAccountHandler.TestAccount)

	// External integration endpoints
	protectedGroup.POST("/projects/:id/external/connect", externalIntegrationHandler.ConnectProjectToAccount)
	protectedGroup.DELETE("/projects/:id/external/disconnect", externalIntegrationHandler.DisconnectProject)
	protectedGroup.GET("/projects/:id/descriptions/suggest", externalIntegrationHandler.GetDescriptionSuggestions)
	protectedGroup.GET("/projects/:id/issues/resolve", externalIntegrationHandler.ResolveIssue)
	protectedGroup.POST("/projects/:id/external/sync", externalIntegrationHandler.SyncProjectIssues)
	protectedGroup.POST("/external/resolve-pending", externalIntegrationHandler.ResolvePendingReferences)
	protectedGroup.GET("/external/issues/:id", externalIntegrationHandler.GetExternalIssue)

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
