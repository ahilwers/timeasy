package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tbaehler/gin-keycloak/pkg/ginkeycloak"
)

type ProjectController struct {
	projectService ProjectService
}

func (projectController *ProjectController) Init(projectService *ProjectService) {
	projectController.projectService = *projectService
}

func (projectController *ProjectController) AddProject(context *gin.Context) {
	var prj Project
	if err := context.ShouldBindJSON(&prj); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ginToken, _ := context.Get("token")
	token := ginToken.(ginkeycloak.KeyCloakToken)
	prj.UserId = token.Sub

	createdProject, err := projectController.projectService.AddProject(&prj)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, createdProject)
}
