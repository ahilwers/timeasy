package http

import (
	"errors"
	"net/http"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type TimeEntryHandler interface {
	AddTimeEntry(context *gin.Context)
}

type timeEntryHandler struct {
	usecase usecase.TimeEntryUsecase
}

func NewTimeEntryHandler(entryUsecase usecase.TimeEntryUsecase) TimeEntryHandler {
	return &timeEntryHandler{
		usecase: entryUsecase,
	}
}

type timeEntryDto struct {
	Id               uuid.UUID `json:"id"`
	Description      string    `json:"description" binding:"required"`
	StartTimeUTCUnix int64     `json:"startTimeUTCUnix" binding:"required"`
	EndTimeUTCUnix   int64
	ProjectId        uuid.UUID `json:"projectId" binding:"required"`
}

func (handler *timeEntryHandler) AddTimeEntry(context *gin.Context) {
	var entryDto timeEntryDto
	if err := context.ShouldBindJSON(&entryDto); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newEntry := handler.createEntryFromDto(entryDto, userId)

	err = handler.usecase.AddTimeEntry(&newEntry)
	if err != nil {
		errorCode := http.StatusInternalServerError
		var userNotFoundError *usecase.UserNotFoundError
		var projectNotFoundError *usecase.ProjectNotFoundError

		switch {
		case errors.As(err, &userNotFoundError):
			errorCode = http.StatusBadRequest
		case errors.As(err, &projectNotFoundError):
			errorCode = http.StatusBadRequest
		default:
			errorCode = http.StatusInternalServerError
		}
		context.JSON(errorCode, gin.H{"error": err.Error()})
		return
	}
	entryDto.Id = newEntry.ID
	context.JSON(http.StatusOK, entryDto)
}

func (handler *timeEntryHandler) createEntryFromDto(dto timeEntryDto, userId uuid.UUID) model.TimeEntry {
	startTime := handler.convertUnitxTimeToTime(dto.StartTimeUTCUnix)
	endTime := handler.convertUnitxTimeToTime(dto.EndTimeUTCUnix)
	return model.TimeEntry{
		ID:          dto.Id,
		Description: dto.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		ProjectId:   dto.ProjectId,
		UserId:      userId,
	}
}

func (handler *timeEntryHandler) convertUnitxTimeToTime(unixTime int64) time.Time {
	var result time.Time
	if unixTime > 0 {
		result = time.Unix(unixTime, 0).UTC()
	}
	return result
}
