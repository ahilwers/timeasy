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
	usecase     usecase.TeamUsecase
	userUsecase usecase.UserUsecase
}

func NewTeamHandler(usecase usecase.TeamUsecase, userUsecase usecase.UserUsecase) TeamHandler {
	return &teamHandler{
		usecase:     usecase,
		userUsecase: userUsecase,
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
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user, err := handler.userUsecase.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	team := handler.createTeamFromDto(teamDto)

	err = handler.usecase.AddTeam(&team, user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"id": team.ID})
}

func (handler *teamHandler) UpdateTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	var teamDto teamInputDto
	if err := context.ShouldBindJSON(&teamDto); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	handler.fillTeamDataFromDto(team, teamDto)

	err = handler.usecase.UpdateTeam(team)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("team %v updated", team.ID)})
}

func (handler *teamHandler) GetTeamById(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	context.JSON(http.StatusOK, handler.createDtoFromTeam(team))
}

func (handler *teamHandler) GetAllTeams(context *gin.Context) {
	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	isAdmin, err := TokenHasRole(tokenString, model.RoleAdmin)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var dtos []teamDto
	if isAdmin {
		teams, err := handler.usecase.GetAllTeams()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		dtos = handler.convertTeamsToDtos(teams)
	} else {
		teamAssignments, err := handler.usecase.GetTeamsOfUser(userId)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, assignment := range teamAssignments {
			dto := handler.createDtoFromTeam(&assignment.Team)
			dtos = append(dtos, dto)
		}
	}

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all teams"})
		return
	}
	context.JSON(http.StatusOK, dtos)
}

func (handler *teamHandler) DeleteTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	err = handler.usecase.DeleteTeam(teamId)
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("team %v deleted", teamId)})
}

type addUserInput struct {
	Id    uuid.UUID      `json:"id" binding:"required"`
	Roles model.RoleList `json:"roles"`
}

func (handler *teamHandler) AddUserToTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var userInput addUserInput
	if err := context.ShouldBindJSON(&userInput); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userToBeAdded, err := handler.userUsecase.GetUserById(userInput.Id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user, err := handler.userUsecase.GetUserById(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !handler.usecase.IsUserAdminInTeam(user, team) {
		context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to add users to this team"})
		return
	}
	_, err = handler.usecase.AddUserToTeam(userToBeAdded, team, userInput.Roles)
	if err != nil {
		var assignmentExistsError *usecase.EntityExistsError
		errorCode := 0
		switch {
		case errors.As(err, &assignmentExistsError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %v added to team %v", userInput.Id, teamId)})
}

func (handler *teamHandler) DeleteUserFromTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	tokenString := ExtractToken(context)
	loggedInUserId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	loggedInUser, err := handler.userUsecase.GetUserById(loggedInUserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !handler.usecase.IsUserAdminInTeam(loggedInUser, team) {
		context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to add users to this team"})
		return
	}

	userIdToBeDeleted, err := handler.getIdParamValue(context, "userId")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userToBeDeleted, err := handler.userUsecase.GetUserById(userIdToBeDeleted)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	err = handler.usecase.DeleteUserFromTeam(userToBeDeleted, team)
	if err != nil {
		var entityNotFoundError *usecase.EntityNotFoundError
		errorCode := 0
		switch {
		case errors.As(err, &entityNotFoundError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("user %v deleted from team %v", userToBeDeleted.ID, teamId)})
}

type teamRolesInput struct {
	Roles model.RoleList `json:"roles" binding:"required"`
}

func (handler *teamHandler) UpdateUserRolesInTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var rolesInput teamRolesInput
	if err := context.ShouldBindJSON(&rolesInput); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userToBeUpdatedId, err := handler.getIdParamValue(context, "userId")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userToBeUpdated, err := handler.userUsecase.GetUserById(userToBeUpdatedId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	tokenString := ExtractToken(context)
	loggedInUserId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	loggedInUser, err := handler.userUsecase.GetUserById(loggedInUserId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !handler.usecase.IsUserAdminInTeam(loggedInUser, team) {
		context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update users in this team"})
		return
	}
	err = handler.usecase.UpdateUserRolesInTeam(userToBeUpdated, team, rolesInput.Roles)
	if err != nil {
		var entityNotFoundError *usecase.EntityNotFoundError
		errorCode := 0
		switch {
		case errors.As(err, &entityNotFoundError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("roles of user %v in team %v updated", userToBeUpdatedId, teamId)})
}

func (handler *teamHandler) getId(context *gin.Context) (uuid.UUID, error) {
	return handler.getIdParamValue(context, "id")
}

func (handler *teamHandler) getIdParamValue(context *gin.Context, paramName string) (uuid.UUID, error) {
	id := context.Param(paramName)
	if id == "" {
		return uuid.Nil, fmt.Errorf("please specify a valid id")
	}
	userId, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, err
	}
	return userId, nil
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
