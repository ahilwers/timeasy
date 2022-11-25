package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tbaehler/gin-keycloak/pkg/ginkeycloak"
)

type ProjectController interface {
	AddProject(context *gin.Context)
}

type projectController struct {
	projectService ProjectService
}

func NewController(projectService ProjectService) ProjectController {
	return &projectController{projectService}
}

func (controller *projectController) AddProject(context *gin.Context) {
	var prj Project
	if err := context.ShouldBindJSON(&prj); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ginToken, _ := context.Get("token")
	token := ginToken.(ginkeycloak.KeyCloakToken)
	prj.UserId = token.Sub

	createdProject, err := controller.projectService.AddProject(&prj)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, createdProject)
}
