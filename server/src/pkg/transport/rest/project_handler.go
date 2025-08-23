package rest

import (
	"fmt"
	"net/http"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
	"timeasy-server/pkg/usecase"

	"github.com/shopspring/decimal"

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
	tokenVerifier          TokenVerifier
	usecase                usecase.ProjectUsecase
	teamUsecase            usecase.TeamUsecase
	externalConnectionRepo repository.ExternalConnectionRepository
}

func NewProjectHandler(tokenVerifier TokenVerifier, usecase usecase.ProjectUsecase, teamUsecase usecase.TeamUsecase, externalConnectionRepo repository.ExternalConnectionRepository) ProjectHandler {
	return &projectHandler{
		tokenVerifier:          tokenVerifier,
		usecase:                usecase,
		teamUsecase:            teamUsecase,
		externalConnectionRepo: externalConnectionRepo,
	}
}

type projectInput struct {
	Name       string          `json:"name" binding:"required"`
	Color      string          `json:"color" binding:"required"`
	Deadline   *model.DateOnly `json:"deadline,omitempty"`
	HourlyRate *float32        `json:"hourlyRate,omitempty"`
	TimeBudget *int            `json:"timeBudget,omitempty"`
	IsActive   *bool           `json:"isActive,omitempty"`
}

type projectTeamAssignmentInput struct {
	ProjectId uuid.UUID `json:"projectId" binding:"required"`
	TeamId    uuid.UUID `json:"teamId" binding:"required"`
}

func (handler *projectHandler) AddProject(context *gin.Context) {
	var prj projectInput
	if err := context.ShouldBindJSON(&prj); err != nil {
		LogHandlerError("AddProject", err, "could not bind project data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("AddProject", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("AddProject", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get optional clientId from URL query parameter
	clientId := context.Query("clientId")

	newProject := model.Project{
		UserId: userId,
	}
	err = handler.fillProjectFromDto(&newProject, prj)
	if err != nil {
		LogHandlerError("AddProject", err, "failed to fill project from DTO")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = handler.usecase.AddProject(&newProject, userId, clientId)
	if err != nil {
		LogHandlerError("AddProject", err, "failed to add project")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, prj)
}

func (handler *projectHandler) UpdateProject(context *gin.Context) {
	projectId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("UpdateProject", err, "failed to get project ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		LogHandlerError("UpdateProject", err, fmt.Sprintf("failed to get project by ID: %v", projectId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
		return
	}

	var prj projectInput
	if err := context.ShouldBindJSON(&prj); err != nil {
		LogHandlerError("UpdateProject", err, "could not bind project data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("UpdateProject", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("UpdateProject", err, "failed to get user ID from token")
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
			LogHandlerError("UpdateProject", err, "failed to check admin role")
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update this project"})
			return
		}
	}

	// Get optional clientId from URL query parameter
	clientId := context.Query("clientId")

	err = handler.fillProjectFromDto(project, prj)
	if err != nil {
		LogHandlerError("UpdateProject", err, "failed to fill project from DTO")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = handler.usecase.UpdateProject(project, userId, clientId)
	if err != nil {
		LogHandlerError("UpdateProject", err, "failed to update project")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, prj)
}

func (handler *projectHandler) DeleteProject(context *gin.Context) {
	projectId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("DeleteProject", err, "failed to get project ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("DeleteProject", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("DeleteProject", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get optional clientId from URL query parameter
	clientId := context.Query("clientId")

	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		LogHandlerError("DeleteProject", err, fmt.Sprintf("failed to get project by ID: %v", projectId))
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
			LogHandlerError("DeleteProject", err, "failed to check admin role")
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
			return
		}
	}
	err = handler.usecase.DeleteProject(projectId, userId, clientId)
	if err != nil {
		LogHandlerError("DeleteProject", err, "failed to delete project")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("project %v deleted", projectId)})
}

func (handler *projectHandler) fillProjectFromDto(project *model.Project, dto projectInput) error {
	project.Name = dto.Name
	project.Color = dto.Color

	if dto.TimeBudget != nil {
		project.TimeBudget = *dto.TimeBudget
	}

	if dto.HourlyRate != nil {
		project.HourlyRate = decimal.NewFromFloat32(*dto.HourlyRate)
	}

	if dto.Deadline != nil {
		if dto.Deadline.IsZero() {
			project.Deadline = model.DateOnly{}
		} else {
			project.Deadline = *dto.Deadline
		}
	}

	if dto.IsActive != nil {
		project.IsActive = *dto.IsActive
	}
	return nil
}

func (handler *projectHandler) GetProjectById(context *gin.Context) {
	projectId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("GetProjectById", err, "failed to get project ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	project, err := handler.usecase.GetProjectById(projectId)
	if err != nil {
		LogHandlerError("GetProjectById", err, fmt.Sprintf("failed to get project by ID: %v", projectId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found", projectId)})
		return
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("GetProjectById", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("GetProjectById", err, "failed to get user ID from token")
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
			LogHandlerError("GetProjectById", err, "failed to check admin role")
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

	// Fetch external connection information
	externalConnection, err := handler.externalConnectionRepo.GetByProjectID(projectId)
	if err != nil {
		LogHandlerError("GetProjectById", err, fmt.Sprintf("Failed to fetch external connection for project %v", projectId))
	} else if externalConnection != nil {
		project.ExternalConnection = externalConnection
		fmt.Printf("DEBUG: Found external connection for project %v: %+v\n", projectId, externalConnection)
	} else {
		fmt.Printf("DEBUG: No external connection found for project %v\n", projectId)
	}
	// Note: We ignore errors here because not all projects have external connections

	context.JSON(http.StatusOK, project)
}

func (handler *projectHandler) GetAllProjects(context *gin.Context) {
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("GetAllProjects", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	hasAdminRole, err := token.HasRole(model.RoleAdmin)
	if err != nil {
		LogHandlerError("GetAllProjects", err, "failed to check admin role")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("GetAllProjects", err, "failed to get user ID from token")
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
		LogHandlerError("GetAllProjects", err, "failed to get projects")
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all project"})
		return
	}
	context.JSON(http.StatusOK, projects)
}

func (handler *projectHandler) AssignProjectToTeam(context *gin.Context) {
	var projectTeamAssignment projectTeamAssignmentInput
	if err := context.ShouldBindJSON(&projectTeamAssignment); err != nil {
		LogHandlerError("AssignProjectToTeam", err, "could not bind project team assignment data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("AssignProjectToTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("AssignProjectToTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get optional clientId from URL query parameter
	clientId := context.Query("clientId")

	project, err := handler.usecase.GetProjectById(projectTeamAssignment.ProjectId)
	if err != nil {
		LogHandlerError("AssignProjectToTeam", err, fmt.Sprintf("failed to get project by ID: %v", projectTeamAssignment.ProjectId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("project with id %v not found",
			projectTeamAssignment.ProjectId)})
		return
	}

	team, err := handler.teamUsecase.GetTeamById(projectTeamAssignment.TeamId)
	if err != nil {
		LogHandlerError("AssignProjectToTeam", err, fmt.Sprintf("failed to get team by ID: %v", projectTeamAssignment.TeamId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found",
			projectTeamAssignment.TeamId)})
		return
	}

	if !handler.teamUsecase.IsUserAdminInTeam(userId, team.ID) {
		isAdmin, err := token.HasRole(model.RoleAdmin)
		if err != nil {
			LogHandlerError("AssignProjectToTeam", err, "failed to check admin role")
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update this project"})
			return
		}
	}

	err = handler.usecase.AssignProjectToTeam(project, team, userId, clientId)
	if err != nil {
		LogHandlerError("AssignProjectToTeam", err, "failed to assign project to team")
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
