package projects

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	//Todo: set user id here

	createdProject, err := controller.projectService.AddProject(&prj)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, createdProject)
}
