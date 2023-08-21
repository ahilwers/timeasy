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
	AssignProjectToTeam(context *gin.Context)
}

type projectHandler struct {
	tokenVerifier TokenVerifier
	usecase       usecase.ProjectUsecase
	teamUsecase   usecase.TeamUsecase
}

func NewProjectHandler(tokenVerifier TokenVerifier, usecase usecase.ProjectUsecase, teamUsecase usecase.TeamUsecase) ProjectHandler {
	return &projectHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
		teamUsecase:   teamUsecase,
	}
}

type projectInput struct {
	Name string `json:"name" binding:"required"`
}

type projectTeamAssignmentInput struct {
	ProjectId uuid.UUID `json:"projectId" binding:"required"`
	TeamId    uuid.UUID `json:"teamId" binding:"required"`
}

func (handler *projectHandler) AddProject(context *gin.Context) {
	var prj projectInput
	if err := context.ShouldBindJSON(&prj); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
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
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// A project belongs to a user if it directly belongs to this user or it belongs to a team the user is member of:
	projectBelongsToUser := userId == project.UserId
	if !projectBelongsToUser && project.TeamID != nil {
		projectBelongsToUser = handler.teamUsecase.IsUserAdminInTeam(userId, *project.TeamID)
	}

	if !projectBelongsToUser {
		isAdmin, err := token.HasRole(model.RoleAdmin)
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
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// A project belongs to a user if it directly belongs to this user or it belongs to a team the user is member of:
	projectBelongsToUser := authUserId == project.UserId
	if !projectBelongsToUser && project.TeamID != nil {
		projectBelongsToUser = handler.teamUsecase.DoesUserBelongToTeam(authUserId, *project.TeamID)
	}

	// a normal user can only fetch his own data.
	// if he tries to get the project of another user he must be an admin.
	if !projectBelongsToUser {
		hasAdminRole, err := token.HasRole(model.RoleAdmin)
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
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	hasAdminRole, err := token.HasRole(model.RoleAdmin)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
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

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
		return
	}

	// A project belongs to a user if it directly belongs to this user or it belongs to a team the user is member of:
	projectBelongsToUser := userId == project.UserId
	if !projectBelongsToUser && project.TeamID != nil {
		projectBelongsToUser = handler.teamUsecase.IsUserAdminInTeam(userId, *project.TeamID)
	}

	if !projectBelongsToUser {
		isAdmin, err := token.HasRole(model.RoleAdmin)
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

func (handler *projectHandler) AssignProjectToTeam(context *gin.Context) {
	var projectTeamAssignment projectTeamAssignmentInput
	if err := context.ShouldBindJSON(&projectTeamAssignment); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userId, err := token.GetUserId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	project, err := handler.usecase.GetProjectById(projectTeamAssignment.ProjectId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found",
			projectTeamAssignment.ProjectId)})
		return
	}

	team, err := handler.teamUsecase.GetTeamById(projectTeamAssignment.TeamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found",
			projectTeamAssignment.TeamId)})
		return
	}

	if !handler.teamUsecase.IsUserAdminInTeam(userId, team.ID) {
		isAdmin, err := token.HasRole(model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update this project"})
			return
		}
	}

	err = handler.usecase.AssignProjectToTeam(project, team)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, project)
}

func (handler *projectHandler) getId(context *gin.Context) (uuid.UUID, error) {
	return handler.getIdParam(context, "id")
}

func (handler *projectHandler) getIdParam(context *gin.Context, paramName string) (uuid.UUID, error) {
	id := context.Param(paramName)
	if id == "" {
		return uuid.Nil, fmt.Errorf("please specify a valid %v", paramName)
	}
	userId, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
}
