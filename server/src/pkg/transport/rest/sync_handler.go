package rest

import (
	"net/http"
	"strconv"
	"time"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
)

type SyncHandler interface {
	GetChangedTimeEntries(context *gin.Context)
	SendLocallyChangedTimeEntries(context *gin.Context)
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

func (handler *syncHandler) SendLocallyChangedTimeEntries(context *gin.Context) {
}

func (handler *syncHandler) GetChangedProjects(context *gin.Context) {
}

func (handler *syncHandler) SendLocallyChangedProjects(context *gin.Context) {
}
