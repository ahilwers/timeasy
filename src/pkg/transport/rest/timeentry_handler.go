package rest

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
	GetTimeEntryById(context *gin.Context)
	GetAllTimeEntries(context *gin.Context)
}

type timeEntryHandler struct {
	usecase usecase.TimeEntryUsecase
}

func NewTimeEntryHandler(entryUsecase usecase.TimeEntryUsecase) TimeEntryHandler {
	return &timeEntryHandler{
		usecase: entryUsecase,
	}
}

type timeEntryUpdateDto struct {
	Description      string `json:"description" binding:"required"`
	StartTimeUTCUnix int64  `json:"startTimeUTCUnix" binding:"required"`
	EndTimeUTCUnix   int64
	ProjectId        uuid.UUID `json:"projectId" binding:"required"`
}

type timeEntryDto struct {
	Id uuid.UUID
	timeEntryUpdateDto
}

func (handler *timeEntryHandler) AddTimeEntry(context *gin.Context) {
	var entryDto timeEntryUpdateDto
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
	var entryDto timeEntryUpdateDto
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

	if timeEntry.UserId != userId {
		isAdmin, err := TokenHasRole(tokenString, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !isAdmin {
			context.JSON(http.StatusForbidden, gin.H{"error": "you are not allowed to update this entry"})
			return
		}
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

func (handler *timeEntryHandler) GetTimeEntryById(context *gin.Context) {
	entryId, err := handler.getId(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	timeEntry, err := handler.usecase.GetTimeEntryById(entryId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("entry with id %v not found", entryId)})
		return
	}
	token := ExtractToken(context)
	authUserId, err := ExtractTokenUserId(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// a normal user can only fetch his own data.
	// if he tries to get an entry of another user he must be an admin.
	if authUserId != timeEntry.UserId {
		hasAdminRole, err := TokenHasRole(token, model.RoleAdmin)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !hasAdminRole {
			// We just say that the entry was not found:
			context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("entry with id %v not found",
				entryId)})
			return
		}
	}
	context.JSON(http.StatusOK, handler.createDtoFromTimeEntry(timeEntry))
}

func (handler *timeEntryHandler) GetAllTimeEntries(context *gin.Context) {
	token := ExtractToken(context)
	userId, err := ExtractTokenUserId(token)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var timeEntries []model.TimeEntry
	timeEntries, err = handler.usecase.GetAllTimeEntriesOfUser(userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all entries"})
		return
	}
	timeEntryDtos := handler.convertTimeEntriesToDtos(timeEntries)
	context.JSON(http.StatusOK, timeEntryDtos)
}

func (handler *timeEntryHandler) createEntryFromDto(dto timeEntryUpdateDto, userId uuid.UUID) model.TimeEntry {
	timeEntry := model.TimeEntry{
		UserId: userId,
	}
	handler.fillEntryFromDto(&timeEntry, dto)
	return timeEntry
}

func (handler *timeEntryHandler) fillEntryFromDto(entry *model.TimeEntry, dto timeEntryUpdateDto) {
	startTime := handler.convertUnixTimeToTime(dto.StartTimeUTCUnix)
	endTime := handler.convertUnixTimeToTime(dto.EndTimeUTCUnix)
	entry.Description = dto.Description
	entry.StartTime = startTime
	entry.EndTime = endTime
	entry.ProjectId = dto.ProjectId
}

func (handler *timeEntryHandler) convertTimeEntriesToDtos(timeEntries []model.TimeEntry) []timeEntryDto {
	var dtos []timeEntryDto
	for _, timeEntry := range timeEntries {
		dtos = append(dtos, handler.createDtoFromTimeEntry(&timeEntry))
	}
	return dtos
}

func (handler *timeEntryHandler) createDtoFromTimeEntry(timeEntry *model.TimeEntry) timeEntryDto {
	dto := timeEntryDto{
		Id: timeEntry.ID,
	}
	dto.Description = timeEntry.Description
	dto.StartTimeUTCUnix = handler.convertTimeToUnixTime(timeEntry.StartTime)
	dto.EndTimeUTCUnix = handler.convertTimeToUnixTime(timeEntry.EndTime)
	dto.ProjectId = timeEntry.ProjectId
	return dto
}

func (handler *timeEntryHandler) convertUnixTimeToTime(unixTime int64) time.Time {
	var result time.Time
	if unixTime > 0 {
		result = time.Unix(unixTime, 0).UTC()
	}
	return result
}

func (handler *timeEntryHandler) convertTimeToUnixTime(time time.Time) int64 {
	result := int64(0)
	if !time.IsZero() {
		result = time.Unix()
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
