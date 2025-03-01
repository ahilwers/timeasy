package rest

import (
	"log"
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
			Id:              entry.ID,
			Description:     entry.Description,
			StartTime:       entry.StartTime.Format(time.RFC3339),
			ProjectId:       entry.ProjectId,
			ChangeType:      changeType,
			ChangeTimestamp: changeTime.Format(time.RFC3339),
		}
		if !entry.EndTime.IsZero() {
			syncTimeEntry.EndTime = entry.EndTime.Format(time.RFC3339)
		}
		syncEntries.TimeEntries = append(syncEntries.TimeEntries, syncTimeEntry)
	}

	projects, err := handler.syncUsecase.GetChangedProjects(userId, time.Unix(unixTime, 0))
	for _, project := range projects {
		changeType := CHANGED
		changeTime := project.UpdatedAt
		if !project.DeletedAt.Time.IsZero() {
			changeType = DELETED
			changeTime = project.DeletedAt.Time
		} else if project.CreatedAt == project.UpdatedAt {
			changeType = NEW
			changeTime = project.CreatedAt
		}
		syncProject := ChangedProjectDto{
			Id:              project.ID,
			Name:            project.Name,
			ChangeType:      changeType,
			ChangeTimestamp: changeTime.Format(time.RFC3339),
		}
		syncEntries.Projects = append(syncEntries.Projects, syncProject)
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
	handler.fillInClientSideChangedProjects(&syncData, syncDtos.Projects, userId)
	handler.fillInClientSideChangedTimeEntries(&syncData, syncDtos.TimeEntries, userId)

	err = handler.syncUsecase.UpdateAndDeleteData(syncData)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (handler *syncHandler) fillInClientSideChangedProjects(syncData *model.SyncData, changedProjects []ChangedProjectDto, userId uuid.UUID) {
	for _, changedProject := range changedProjects {
		project := handler.createProjectFromDto(changedProject, userId)
		switch changedProject.ChangeType {
		case NEW, CHANGED:
			syncData.ProjectsToBeUpdated = append(syncData.ProjectsToBeUpdated, project)
		case DELETED:
			syncData.ProjectsToBeDeleted = append(syncData.ProjectsToBeDeleted, project)
		}
	}
}

func (handle *syncHandler) createProjectFromDto(projectDto ChangedProjectDto, userId uuid.UUID) model.Project {
	project := model.Project{
		ID:     projectDto.Id,
		Name:   projectDto.Name,
		UserId: userId,
	}
	return project
}

func (handler *syncHandler) fillInClientSideChangedTimeEntries(syncData *model.SyncData, changedTimeEntries []ChangedTimeEntryDto, userId uuid.UUID) {
	for _, changedTimeEntry := range changedTimeEntries {
		timeEntry, err := handler.createTimeEntryFromDto(changedTimeEntry, userId)
		if err == nil {
			switch changedTimeEntry.ChangeType {
			case NEW, CHANGED:
				syncData.TimeEntriesToBeUpdated = append(syncData.TimeEntriesToBeUpdated, timeEntry)
			case DELETED:
				syncData.TimeEntriesToBeDeleted = append(syncData.TimeEntriesToBeDeleted, timeEntry)
			}
		} else {
			log.Printf("Could not create time entry from dto: %v\n", err)
		}
	}
}

func (handler *syncHandler) createTimeEntryFromDto(timeEntryDto ChangedTimeEntryDto, userId uuid.UUID) (model.TimeEntry, error) {
	startTime, err := time.Parse(time.RFC3339, timeEntryDto.StartTime)
	if err != nil {
		return model.TimeEntry{}, err
	}
	timeEntry := model.TimeEntry{
		ID:          timeEntryDto.Id,
		ProjectId:   timeEntryDto.ProjectId,
		UserId:      userId,
		Description: timeEntryDto.Description,
		StartTime:   startTime,
	}
	if timeEntryDto.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, timeEntryDto.EndTime)
		if err != nil {
			return model.TimeEntry{}, err
		}
		timeEntry.EndTime = endTime
	}
	return timeEntry, nil
}
