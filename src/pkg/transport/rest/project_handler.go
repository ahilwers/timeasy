package rest

import (
	"fmt"
	"net/http"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type ProjectHandler interface {
	AddProject(context *gin.Context)
	GetProjectById(context *gin.Context)
	GetAllProjects(context *gin.Context)
	UpdateProject(context *gin.Context)
	DeleteProject(context *gin.Context)
}

type projectHandler struct {
	usecase usecase.ProjectUsecase
}

func NewProjectHandler(usecase usecase.ProjectUsecase) ProjectHandler {
	return &projectHandler{
		usecase: usecase,
	}
}

type projectInput struct {
	Name string `json:"name" binding:"required"`
}

func (handler *projectHandler) AddProject(context *gin.Context) {
	var prj projectInput
	if err := context.ShouldBindJSON(&prj); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newProject := model.Project{
		Name:   prj.Name,
		UserId: userId,
	}

	err = handler.usecase.AddProject(&newProject)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, prj)
}

func (handler *projectHandler) UpdateProject(context *gin.Context) {
	projectId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
		return
	}

	var prj projectInput
	if err := context.ShouldBindJSON(&prj); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if project.UserId != userId {
		isAdmin, err := TokenHasRole(tokenString, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update this project"})
			return
		}
	}

	project.Name = prj.Name

	err = handler.usecase.UpdateProject(project)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, prj)
}

func (handler *projectHandler) GetProjectById(context *gin.Context) {
	projectId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
		return
	}
	token := ExtractToken(context)
	authUserId, err := ExtractTokenUserId(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// a normal user can only fetch his own data.
	// if he tries to get the project of another user he must be an admin.
	if authUserId != project.UserId {
		hasAdminRole, err := TokenHasRole(token, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !hasAdminRole {
			// We just say that the project was not found:
			context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found",
				projectId)})
			return
		}
	}
	context.JSON(http.StatusOK, project)
}

func (handler *projectHandler) GetAllProjects(context *gin.Context) {
	token := ExtractToken(context)
	hasAdminRole, err := TokenHasRole(token, model.RoleAdmin)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userId, err := ExtractTokenUserId(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var projects []model.Project
	if hasAdminRole {
		projects, err = handler.usecase.GetAllProjects()
	} else {
		projects, err = handler.usecase.GetAllProjectsOfUser(userId)
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all project"})
		return
	}
	context.JSON(http.StatusOK, projects)
}

func (handler *projectHandler) DeleteProject(context *gin.Context) {
	projectId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
		return
	}
	if project.UserId != userId {
		isAdmin, err := TokenHasRole(tokenString, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
			return
		}
	}
	err = handler.usecase.DeleteProject(projectId)
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("project %v deleted", projectId)})
}

func (handler *projectHandler) getId(context *gin.Context) (uuid.UUID, error) {
	id := context.Param("id")
	if id == "" {
		return uuid.Nil, fmt.Errorf("please specify a valid id")
	}
	userId, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}
