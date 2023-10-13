package rest

import (
	"net/http"
	"strconv"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type SyncHandler interface {
	GetChangedTimeEntries(context *gin.Context)
	SendLocallyChangedEntries(context *gin.Context)
	GetChangedProjects(context *gin.Context)
	SendLocallyChangedProjects(context *gin.Context)
}

type syncHandler struct {
	tokenVerifier    TokenVerifier
	timeEntryUsecase usecase.TimeEntryUsecase
}

func NewSyncHandler(tokenVerifier TokenVerifier, timeEntryUsecase usecase.TimeEntryUsecase) SyncHandler {
	return &syncHandler{
		tokenVerifier:    tokenVerifier,
		timeEntryUsecase: timeEntryUsecase,
	}
}

func (handler *syncHandler) GetChangedTimeEntries(context *gin.Context) {
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
	timeParam := context.Param("timestamp")
	unixTime, err := strconv.ParseInt(timeParam, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "please provide a valid unix timestamp"})
		return
	}

	var syncEntries []ChangedTimeEntryDto
	entries, err := handler.timeEntryUsecase.GetChangedEntries(userId, time.Unix(unixTime, 0))
	for _, entry := range entries {
		changeType := CHANGED
		changeTime := entry.UpdatedAt
		if !entry.DeletedAt.Time.IsZero() {
			changeType = DELETED
			changeTime = entry.DeletedAt.Time
		} else if entry.CreatedAt == entry.UpdatedAt {
			changeType = NEW
			changeTime = entry.CreatedAt
		}
		syncEntry := ChangedTimeEntryDto{
			Id:                     entry.ID,
			Description:            entry.Description,
			StartTimeUTCUnix:       entry.StartTime.Unix(),
			EndTimeUTCUnix:         entry.EndTime.Unix(),
			ProjectId:              entry.ProjectId,
			ChangeType:             changeType,
			ChangeTimestampUTCUnix: changeTime.Unix(),
		}
		syncEntries = append(syncEntries, syncEntry)
	}
	context.JSON(http.StatusOK, syncEntries)
}

func (handler *syncHandler) SendLocallyChangedEntries(context *gin.Context) {
	var syncDtos SyncEntries
	if err := context.ShouldBindJSON(&syncDtos); err != nil {
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

	err = handler.handleClientSideChangedTimeEntries(syncDtos.TimeEntries, userId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (handler *syncHandler) handleClientSideChangedTimeEntries(changedTimeEntries []ChangedTimeEntryDto, userId uuid.UUID) error {
	var timeEntriesToBeUpdated []model.TimeEntry
	var timeEntryIdsToBeDeleted []uuid.UUID
	for _, changedTimeEntry := range changedTimeEntries {
		if changedTimeEntry.ChangeType == DELETED {
			timeEntryIdsToBeDeleted = append(timeEntryIdsToBeDeleted, changedTimeEntry.Id)
		} else {
			timeEntry := handler.createTimeEntryFromDto(changedTimeEntry, userId)
			timeEntriesToBeUpdated = append(timeEntriesToBeUpdated, timeEntry)
		}
	}
	err := handler.timeEntryUsecase.UpdateTimeEntryList(timeEntriesToBeUpdated)
	if err != nil {
		return err
	}
	return nil
}

func (handler *syncHandler) createTimeEntryFromDto(timeEntryDto ChangedTimeEntryDto, userId uuid.UUID) model.TimeEntry {
	timeEntry := model.TimeEntry{
		ID:          timeEntryDto.Id,
		ProjectId:   timeEntryDto.ProjectId,
		UserId:      userId,
		Description: timeEntryDto.Description,
		StartTime:   time.Unix(timeEntryDto.StartTimeUTCUnix, 0).UTC(),
		EndTime:     time.Unix(timeEntryDto.EndTimeUTCUnix, 0).UTC(),
	}
	timeEntry.UpdatedAt = time.Unix(timeEntryDto.ChangeTimestampUTCUnix, 0).UTC()
	if timeEntryDto.ChangeType == NEW {
		timeEntry.CreatedAt = time.Unix(timeEntryDto.ChangeTimestampUTCUnix, 0).UTC()
	}
	return timeEntry
}

func (handler *syncHandler) GetChangedProjects(context *gin.Context) {
}

func (handler *syncHandler) SendLocallyChangedProjects(context *gin.Context) {
}
