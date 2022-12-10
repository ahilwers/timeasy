package http

import (
	"net/http"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
)

type ProjectHandler interface {
	AddProject(context *gin.Context)
}

type projectHandler struct {
	usecase usecase.ProjectUsecase
}

func NewProjectHandler(usecase usecase.ProjectUsecase) ProjectHandler {
	return &projectHandler{
		usecase: usecase,
	}
}

func (handler *projectHandler) AddProject(context *gin.Context) {
	var prj model.Project
	if err := context.ShouldBindJSON(&prj); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//Todo: set user id here

	createdProject, err := handler.usecase.AddProject(&prj)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, createdProject)
}
