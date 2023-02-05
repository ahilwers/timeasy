package http

import (
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
	}
	context.JSON(http.StatusOK, gin.H{"id": team.ID})
}

func (handler *teamHandler) UpdateTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	}
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("team %v updated", team.ID)})
}

func (handler *teamHandler) GetTeamById(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	team, err := handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	context.JSON(http.StatusOK, handler.createDtoFromTeam(team))
}

func (handler *teamHandler) GetAllTeams(context *gin.Context) {
	var teams []model.Team
	teams, err := handler.usecase.GetAllTeams()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all teams"})
		return
	}
	context.JSON(http.StatusOK, handler.convertTeamsToDtos(teams))
}

func (handler *teamHandler) DeleteTeam(context *gin.Context) {
	teamId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	_, err = handler.usecase.GetTeamById(teamId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("team with id %v not found", teamId)})
		return
	}
	err = handler.usecase.DeleteTeam(teamId)
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("team %v deleted", teamId)})
}

func (handler *teamHandler) getId(context *gin.Context) (uuid.UUID, error) {
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
