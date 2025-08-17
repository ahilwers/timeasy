package rest

import (
	"errors"
	"fmt"
	"net/http"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type TeamHandler interface {
	AddTeam(context *gin.Context)
	GetTeamById(context *gin.Context)
	GetAllTeams(context *gin.Context)
	UpdateTeam(context *gin.Context)
	DeleteTeam(context *gin.Context)
	AddUserToTeam(context *gin.Context)
	DeleteUserFromTeam(context *gin.Context)
	UpdateUserRolesInTeam(context *gin.Context)
}

type teamHandler struct {
	tokenVerifier TokenVerifier
	usecase       usecase.TeamUsecase
}

func NewTeamHandler(tokenVerifier TokenVerifier, usecase usecase.TeamUsecase) TeamHandler {
	return &teamHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
	}
}

type teamDto struct {
	ID uuid.UUID
	teamInputDto
}

type teamInputDto struct {
	Name1 string `json:"name1" binding:"required"`
	Name2 string `json:"name2"`
	Name3 string `json:"name3"`
}

func (handler *teamHandler) AddTeam(context *gin.Context) {
	var teamDto teamInputDto
	if err := context.ShouldBindJSON(&teamDto); err != nil {
		LogHandlerError("AddTeam", err, "could not bind team data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("AddTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("AddTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clientId := context.Query("clientId")

	team := handler.createTeamFromDto(teamDto)

	err = handler.usecase.AddTeam(&team, userId, clientId)
	if err != nil {
		LogHandlerError("AddTeam", err, "failed to add team")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"id": team.ID})
}

func (handler *teamHandler) UpdateTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("UpdateTeam", err, "failed to get team ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		LogHandlerError("UpdateTeam", err, fmt.Sprintf("failed to get team by ID: %v", teamId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	var teamDto teamInputDto
	if err := context.ShouldBindJSON(&teamDto); err != nil {
		LogHandlerError("UpdateTeam", err, "could not bind team data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("UpdateTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("UpdateTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clientId := context.Query("clientId")

	handler.fillTeamDataFromDto(team, teamDto)

	err = handler.usecase.UpdateTeam(team, userId, clientId)
	if err != nil {
		LogHandlerError("UpdateTeam", err, "failed to update team")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("team %v updated", team.ID)})
}

func (handler *teamHandler) GetTeamById(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("GetTeamById", err, "failed to get team ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		LogHandlerError("GetTeamById", err, fmt.Sprintf("failed to get team by ID: %v", teamId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	context.JSON(http.StatusOK, handler.createDtoFromTeam(team))
}

func (handler *teamHandler) GetAllTeams(context *gin.Context) {
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("GetAllTeams", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("GetAllTeams", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	isAdmin, err := token.HasRole(model.RoleAdmin)
	if err != nil {
		LogHandlerError("GetAllTeams", err, "failed to check admin role")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var dtos []teamDto
	if isAdmin {
		teams, err := handler.usecase.GetAllTeams()
		if err != nil {
			LogHandlerError("GetAllTeams", err, "failed to get all teams")
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		dtos = handler.convertTeamsToDtos(teams)
	} else {
		teamAssignments, err := handler.usecase.GetTeamsOfUser(userId)
		if err != nil {
			LogHandlerError("GetAllTeams", err, "failed to get teams of user")
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, assignment := range teamAssignments {
			dto := handler.createDtoFromTeam(&assignment.Team)
			dtos = append(dtos, dto)
		}
	}
	context.JSON(http.StatusOK, dtos)
}

func (handler *teamHandler) DeleteTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("DeleteTeam", err, "failed to get team ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = handler.usecase.GetTeamById(teamId)
	if err != nil {
		LogHandlerError("DeleteTeam", err, fmt.Sprintf("failed to get team by ID: %v", teamId))
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("DeleteTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("DeleteTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clientId := context.Query("clientId")

	err = handler.usecase.DeleteTeam(teamId, userId, clientId)
	if err != nil {
		LogHandlerError("DeleteTeam", err, "failed to delete team")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("team %v deleted", teamId)})
}

type addUserInput struct {
	Id    uuid.UUID      `json:"id" binding:"required"`
	Roles model.RoleList `json:"roles"`
}

func (handler *teamHandler) AddUserToTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("AddUserToTeam", err, "failed to get team ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		LogHandlerError("AddUserToTeam", err, fmt.Sprintf("failed to get team by ID: %v", teamId))
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var userInput addUserInput
	if err := context.ShouldBindJSON(&userInput); err != nil {
		LogHandlerError("AddUserToTeam", err, "could not bind user input data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("AddUserToTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("AddUserToTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clientId := context.Query("clientId")
	if !handler.usecase.IsUserAdminInTeam(authUserId, team.ID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to add users to this team"})
		return
	}
	_, err = handler.usecase.AddUserToTeam(userInput.Id, team, userInput.Roles, authUserId, clientId)
	if err != nil {
		var assignmentExistsError *usecase.EntityExistsError
		errorCode := 0
		switch {
		case errors.As(err, &assignmentExistsError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		LogHandlerError("AddUserToTeam", err, "failed to add user to team")
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %v added to team %v", userInput.Id, teamId)})
}

func (handler *teamHandler) DeleteUserFromTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("DeleteUserFromTeam", err, "failed to get team ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		LogHandlerError("DeleteUserFromTeam", err, fmt.Sprintf("failed to get team by ID: %v", teamId))
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("DeleteUserFromTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("DeleteUserFromTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !handler.usecase.IsUserAdminInTeam(authUserId, team.ID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to add users to this team"})
		return
	}

	userIdToBeDeleted, err := GetMandatoryIdParamValue(context, "userId")
	if err != nil {
		LogHandlerError("DeleteUserFromTeam", err, "failed to get user ID parameter")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clientId := context.Query("clientId")
	err = handler.usecase.DeleteUserFromTeam(userIdToBeDeleted, team, authUserId, clientId)
	if err != nil {
		var entityNotFoundError *usecase.EntityNotFoundError
		errorCode := 0
		switch {
		case errors.As(err, &entityNotFoundError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		LogHandlerError("DeleteUserFromTeam", err, "failed to delete user from team")
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %v deleted from team %v", userIdToBeDeleted, teamId)})
}

type teamRolesInput struct {
	Roles model.RoleList `json:"roles" binding:"required"`
}

func (handler *teamHandler) UpdateUserRolesInTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		LogHandlerError("UpdateUserRolesInTeam", err, "failed to get team ID")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		LogHandlerError("UpdateUserRolesInTeam", err, fmt.Sprintf("failed to get team by ID: %v", teamId))
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var rolesInput teamRolesInput
	if err := context.ShouldBindJSON(&rolesInput); err != nil {
		LogHandlerError("UpdateUserRolesInTeam", err, "could not bind roles input data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userToBeUpdatedId, err := GetMandatoryIdParamValue(context, "userId")
	if err != nil {
		LogHandlerError("UpdateUserRolesInTeam", err, "failed to get user ID parameter")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("UpdateUserRolesInTeam", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	authUserId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("UpdateUserRolesInTeam", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !handler.usecase.IsUserAdminInTeam(authUserId, team.ID) {
		context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update users in this team"})
		return
	}
	clientId := context.Query("clientId")
	err = handler.usecase.UpdateUserRolesInTeam(userToBeUpdatedId, team, rolesInput.Roles, authUserId, clientId)
	if err != nil {
		var entityNotFoundError *usecase.EntityNotFoundError
		errorCode := 0
		switch {
		case errors.As(err, &entityNotFoundError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		LogHandlerError("UpdateUserRolesInTeam", err, "failed to update user roles in team")
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("roles of user %v in team %v updated", userToBeUpdatedId, teamId)})
}

func (handler *teamHandler) getId(context *gin.Context) (uuid.UUID, error) {
	return GetMandatoryIdParamValue(context, "id")
}

func (handler *teamHandler) createTeamFromDto(dto teamInputDto) model.Team {
	team := model.Team{}
	handler.fillTeamDataFromDto(&team, dto)
	return team
}

func (handler *teamHandler) fillTeamDataFromDto(team *model.Team, dto teamInputDto) {
	team.Name1 = dto.Name1
	team.Name2 = dto.Name2
	team.Name3 = dto.Name3
}

func (handler *teamHandler) convertTeamsToDtos(teams []model.Team) []teamDto {
	var dtos []teamDto
	for _, team := range teams {
		dto := handler.createDtoFromTeam(&team)
		dtos = append(dtos, dto)
	}
	return dtos
}

func (handler *teamHandler) createDtoFromTeam(team *model.Team) teamDto {
	dto := teamDto{
		ID: team.ID,
	}
	dto.Name1 = team.Name1
	dto.Name2 = team.Name2
	dto.Name3 = team.Name3
	return dto
}
