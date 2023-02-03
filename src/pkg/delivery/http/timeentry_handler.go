package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type TimeEntryHandler interface {
	AddTimeEntry(context *gin.Context)
	UpdateTimeEntry(context *gin.Context)
	DeleteTimeEntry(context *gin.Context)
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
	Description      string `json:"description" binding:"required"`
	StartTimeUTCUnix int64  `json:"startTimeUTCUnix" binding:"required"`
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
	context.JSON(http.StatusOK, gin.H{"id": newEntry.ID})
}

func (handler *timeEntryHandler) UpdateTimeEntry(context *gin.Context) {
	entryId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	timeEntry, err := handler.usecase.GetTimeEntryById(entryId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("entry with id %v not found", entryId)})
		return
	}
	var entryDto timeEntryDto
	if err := context.ShouldBindJSON(&entryDto); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenString := ExtractToken(context)
	_, err = ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.fillEntryFromDto(timeEntry, entryDto)

	err = handler.usecase.UpdateTimeEntry(timeEntry)
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
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("entry %v updated", entryId)})
}

func (handler *timeEntryHandler) DeleteTimeEntry(context *gin.Context) {
	entryId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	tokenString := ExtractToken(context)
	userId, err := ExtractTokenUserId(tokenString)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	timeEntry, err := handler.usecase.GetTimeEntryById(entryId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("entry with id %v not found", entryId)})
		return
	}
	if timeEntry.UserId != userId {
		isAdmin, err := TokenHasRole(tokenString, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("entry with id %v not found", entryId)})
			return
		}
	}
	err = handler.usecase.DeleteTimeEntry(entryId)
	context.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("entry %v deleted", entryId)})
}

func (handler *timeEntryHandler) createEntryFromDto(dto timeEntryDto, userId uuid.UUID) model.TimeEntry {
	timeEntry := model.TimeEntry{
		UserId: userId,
	}
	handler.fillEntryFromDto(&timeEntry, dto)
	return timeEntry
}

func (handler *timeEntryHandler) fillEntryFromDto(entry *model.TimeEntry, dto timeEntryDto) {
	startTime := handler.convertUnitxTimeToTime(dto.StartTimeUTCUnix)
	endTime := handler.convertUnitxTimeToTime(dto.EndTimeUTCUnix)
	entry.Description = dto.Description
	entry.StartTime = startTime
	entry.EndTime = endTime
	entry.ProjectId = dto.ProjectId
}

func (handler *timeEntryHandler) convertUnitxTimeToTime(unixTime int64) time.Time {
	var result time.Time
	if unixTime > 0 {
		result = time.Unix(unixTime, 0).UTC()
	}
	return result
}

func (handler *timeEntryHandler) getId(context *gin.Context) (uuid.UUID, error) {
	idParam := context.Param("id")
	if idParam == "" {
		return uuid.Nil, fmt.Errorf("please specify a valid id")
	}
	id, err := uuid.FromString(idParam)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
