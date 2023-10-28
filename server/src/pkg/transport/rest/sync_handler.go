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
	GetChangedEntries(context *gin.Context)
	SendLocallyChangedEntries(context *gin.Context)
	GetChangedProjects(context *gin.Context)
	SendLocallyChangedProjects(context *gin.Context)
}

type syncHandler struct {
	tokenVerifier TokenVerifier
	syncUsecase   usecase.SyncUsecase
}

func NewSyncHandler(tokenVerifier TokenVerifier, syncUsecase usecase.SyncUsecase) SyncHandler {
	return &syncHandler{
		tokenVerifier: tokenVerifier,
		syncUsecase:   syncUsecase,
	}
}

func (handler *syncHandler) GetChangedEntries(context *gin.Context) {
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

	var syncEntries SyncEntries
	entries, err := handler.syncUsecase.GetChangedTimeEntries(userId, time.Unix(unixTime, 0))
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
		syncTimeEntry := ChangedTimeEntryDto{
			Id:                     entry.ID,
			Description:            entry.Description,
			StartTimeUTCUnix:       entry.StartTime.Unix(),
			EndTimeUTCUnix:         entry.EndTime.Unix(),
			ProjectId:              entry.ProjectId,
			ChangeType:             changeType,
			ChangeTimestampUTCUnix: changeTime.Unix(),
		}
		syncEntries.TimeEntries = append(syncEntries.TimeEntries, syncTimeEntry)
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

	var syncData model.SyncData
	handler.fillInClientSideChangedTimeEntries(&syncData, syncDtos.TimeEntries, userId)

	err = handler.syncUsecase.UpdateAndDeleteData(syncData)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (handler *syncHandler) fillInClientSideChangedTimeEntries(syncData *model.SyncData, changedTimeEntries []ChangedTimeEntryDto, userId uuid.UUID) {
	for _, changedTimeEntry := range changedTimeEntries {
		timeEntry := handler.createTimeEntryFromDto(changedTimeEntry, userId)
		switch changedTimeEntry.ChangeType {
		case NEW, CHANGED:
			syncData.TimeEntriesToBeUpdated = append(syncData.TimeEntriesToBeUpdated, timeEntry)
		case DELETED:
			syncData.TimeEntriesToBeDeleted = append(syncData.TimeEntriesToBeDeleted, timeEntry)
		}
	}
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
	return timeEntry
}

func (handler *syncHandler) GetChangedProjects(context *gin.Context) {
}

func (handler *syncHandler) SendLocallyChangedProjects(context *gin.Context) {
}
