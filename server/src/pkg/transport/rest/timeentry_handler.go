package rest

import (
	"errors"
	"fmt"
	"log"
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
	tokenVerifier TokenVerifier
	usecase       usecase.TimeEntryUsecase
}

func NewTimeEntryHandler(tokenVerifier TokenVerifier, entryUsecase usecase.TimeEntryUsecase) TimeEntryHandler {
	return &timeEntryHandler{
		tokenVerifier: tokenVerifier,
		usecase:       entryUsecase,
	}
}

type timeEntryUpdateDto struct {
	Description string    `json:"description,omitempty"`
	StartTime   string    `json:"startTime" binding:"required"`
	EndTime     string    `json:"endTime,omitempty"`
	ProjectId   uuid.UUID `json:"projectId" binding:"required"`
}

type timeEntryDto struct {
	Id string `json:"id" binding:"required"`
	timeEntryUpdateDto
}

func (handler *timeEntryHandler) AddTimeEntry(context *gin.Context) {
	var entryDto timeEntryUpdateDto
	if err := context.ShouldBindJSON(&entryDto); err != nil {
		log.Printf("Could not bind json: %v\n", err)
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

	if timeEntry.UserId != userId {
		isAdmin, err := token.HasRole(model.RoleAdmin)
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

	timeEntry, err := handler.usecase.GetTimeEntryById(entryId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("entry with id %v not found", entryId)})
		return
	}
	if timeEntry.UserId != userId {
		isAdmin, err := token.HasRole(model.RoleAdmin)
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
	// a normal user can only fetch his own data.
	// if he tries to get an entry of another user he must be an admin.
	if authUserId != timeEntry.UserId {
		hasAdminRole, err := token.HasRole(model.RoleAdmin)
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

	projectIdStr := context.Query("projectId")
	startDateStr := context.Query("startDate")
	endDateStr := context.Query("endDate")

	projectId := uuid.Nil
	if projectIdStr != "" {
		projectId, err = uuid.FromString(projectIdStr)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid projectId"})
			return
		}
	}

	if (startDateStr != "" && endDateStr == "") || (startDateStr == "" && endDateStr != "") {
		context.JSON(http.StatusBadRequest, gin.H{"error": "both startDate and endDate must be provided"})
		return
	}

	var startDate time.Time
	var endDate time.Time
	if startDateStr != "" && endDateStr != "" {
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid startDate format, must be RFC3339"})
			return
		}
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid endDate format, must be RFC3339"})
			return
		}
	}

	var timeEntries []model.TimeEntry
	if projectId == uuid.Nil && startDate.IsZero() && endDate.IsZero() {
		timeEntries, err = handler.usecase.GetAllTimeEntriesOfUser(userId)
	} else if projectId != uuid.Nil && startDate.IsZero() && endDate.IsZero() {
		timeEntries, err = handler.usecase.GetAllTimeEntriesOfUserAndProject(userId, projectId)
	} else {
		timeEntries, err = handler.usecase.GetTimeEntriesOfUserAndProjectBetweenDates(userId, projectId, startDate, endDate)
	}
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

func (handler *timeEntryHandler) fillEntryFromDto(entry *model.TimeEntry, dto timeEntryUpdateDto) error {
	entry.Description = dto.Description
	startTime, err := time.Parse(time.RFC3339, dto.StartTime)
	if err != nil {
		return err
	}
	entry.StartTime = startTime
	if dto.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, dto.EndTime)
		if err != nil {
			return err
		}
		entry.EndTime = endTime
	}
	entry.ProjectId = dto.ProjectId
	return nil
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
		Id: timeEntry.ID.String(),
	}
	dto.Description = timeEntry.Description
	dto.StartTime = timeEntry.StartTime.Format(time.RFC3339)
	if !timeEntry.EndTime.IsZero() {
		dto.EndTime = timeEntry.EndTime.Format(time.RFC3339)
	}
	dto.ProjectId = timeEntry.ProjectId
	return dto
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
