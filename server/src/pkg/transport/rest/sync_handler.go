package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
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
		LogHandlerError("GetChangedEntries", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("GetChangedEntries", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	changeLogEntryParam := context.Param("sinceChangeLogEntry")
	sinceChangeLogEntry, err := strconv.ParseInt(changeLogEntryParam, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "please provide a valid id for a changelog entry"})
		return
	}
	clientId := context.Query("clientId")

	var syncEntries SyncEntries
	latestChangeLogEntry, err := handler.syncUsecase.GetLatestChangelogEntryId()
	if err != nil {
		LogHandlerError("GetChangedEntries", err, "failed to get latest changelog entry ID")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	syncEntries.LatestChangeLogId = latestChangeLogEntry

	entries, err := handler.syncUsecase.GetChangedTimeEntries(userId, sinceChangeLogEntry, latestChangeLogEntry, clientId)
	if err != nil {
		LogHandlerError("GetChangedEntries", err, "failed to get changed time entries")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	handler.appendChangedTimeEntries(entries.Created, &syncEntries, NEW)
	handler.appendChangedTimeEntries(entries.Updated, &syncEntries, CHANGED)
	handler.appendChangedTimeEntries(entries.Deleted, &syncEntries, DELETED)

	projects, err := handler.syncUsecase.GetChangedProjects(userId, sinceChangeLogEntry, latestChangeLogEntry, clientId)
	if err != nil {
		LogHandlerError("GetChangedEntries", err, "failed to get changed projects")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	handler.appendChangedProjects(projects.Created, &syncEntries, NEW)
	handler.appendChangedProjects(projects.Updated, &syncEntries, CHANGED)
	handler.appendChangedProjects(projects.Deleted, &syncEntries, DELETED)

	context.JSON(http.StatusOK, syncEntries)
}

func (handler *syncHandler) appendChangedTimeEntries(timeEntries []model.TimeEntry, syncEntries *SyncEntries, changeType ChangeType) {
    for _, entry := range timeEntries {
        // Skip entries with invalid start times
        if entry.StartTime.IsZero() || entry.StartTime.Year() < 1900 {
            LogHandlerError("appendChangedTimeEntries", nil, fmt.Sprintf("skipping time entry %s with invalid start time: %v", entry.ID, entry.StartTime))
            continue
        }
        
        syncTimeEntry := ChangedTimeEntryDto{
            Id:              entry.ID,
            Description:     entry.Description,
            StartTime:       entry.StartTime.UTC().Truncate(time.Second).Format(time.RFC3339),
            ProjectId:       entry.ProjectId,
            ChangeType:      changeType,
            ChangeTimestamp: entry.ChangeAt.UTC().Truncate(time.Second).Format(time.RFC3339),
            ChangeLogId:     entry.ChangeLogId,
        }
		
		// Only include EndTime if it's valid
		if !entry.EndTime.IsZero() && entry.EndTime.Year() >= 1900 {
			syncTimeEntry.EndTime = entry.EndTime.UTC().Truncate(time.Second).Format(time.RFC3339)
		}
		
		syncEntries.TimeEntries = append(syncEntries.TimeEntries, syncTimeEntry)
	}
}

func (handler *syncHandler) appendChangedProjects(projects []model.Project, syncEntries *SyncEntries, changeType ChangeType) {
    for _, project := range projects {
        deadline := project.Deadline
        hourlyRateFloat, _ := project.HourlyRate.Float64()
        timeBudget := project.TimeBudget
        isActive := project.IsActive
        color := project.Color
        
        syncProject := ChangedProjectDto{
            Id:              project.ID,
            Name:            project.Name,
            Color:           &color,
            HourlyRate:      &hourlyRateFloat,
            TimeBudget:      &timeBudget,
            IsActive:        &isActive,
            ChangeType:      changeType,
            ChangeTimestamp: project.ChangeAt.UTC().Truncate(time.Second).Format(time.RFC3339),
            ChangeLogId:     project.ChangeLogId,
        }
		
		// Only set deadline if it's not zero/null
		if !deadline.IsZero() {
			syncProject.Deadline = &deadline
		}
		
		syncEntries.Projects = append(syncEntries.Projects, syncProject)
	}
}

func (handler *syncHandler) parseTimestamp(timestamp int64) time.Time {
	if timestamp > 1e12 {
		return time.UnixMilli(timestamp)
	}
	return time.Unix(timestamp, 0)
}

func (handler *syncHandler) SendLocallyChangedEntries(context *gin.Context) {
	var syncDtos SyncEntries
	if err := context.ShouldBindJSON(&syncDtos); err != nil {
		LogHandlerError("SendLocallyChangedEntries", err, "could not bind sync entries data")
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := handler.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("SendLocallyChangedEntries", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("SendLocallyChangedEntries", err, "failed to get user ID from token")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clientId := context.Query("clientId")

	var syncData model.SyncData
	handler.fillInClientSideChangedProjects(&syncData, syncDtos.Projects, userId)
	handler.fillInClientSideChangedTimeEntries(&syncData, syncDtos.TimeEntries, userId)

	err = handler.syncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
	if err != nil {
		LogHandlerError("SendLocallyChangedEntries", err, "failed to update and delete sync data")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func (handler *syncHandler) fillInClientSideChangedProjects(syncData *model.SyncData, changedProjects []ChangedProjectDto, userId uuid.UUID) {
	for _, changedProject := range changedProjects {
		project := handler.createProjectFromDto(changedProject, userId)
		switch changedProject.ChangeType {
		case NEW:
			syncData.ProjectsToBeCreated = append(syncData.ProjectsToBeCreated, project)
		case CHANGED:
			syncData.ProjectsToBeUpdated = append(syncData.ProjectsToBeUpdated, project)
		case DELETED:
			syncData.ProjectsToBeDeleted = append(syncData.ProjectsToBeDeleted, project)
		}
	}
}

func (handler *syncHandler) createProjectFromDto(projectDto ChangedProjectDto, userId uuid.UUID) model.Project {
	project := model.Project{
		ID:     projectDto.Id,
		UserId: userId,
	}
	existingProject, err := handler.syncUsecase.GetProjectById(projectDto.Id)
	if err == nil && existingProject != nil {
		project = *existingProject
	}

	project.Name = projectDto.Name

	if projectDto.Color != nil {
		project.Color = *projectDto.Color
	}

	if projectDto.TimeBudget != nil {
		project.TimeBudget = *projectDto.TimeBudget
	}

	if projectDto.HourlyRate != nil {
		project.HourlyRate = decimal.NewFromFloat(*projectDto.HourlyRate)
	}

	if projectDto.Deadline != nil {
		if projectDto.Deadline.IsZero() {
			project.Deadline = model.DateOnly{}
		} else {
			project.Deadline = *projectDto.Deadline
		}
	}

	if projectDto.IsActive != nil {
		project.IsActive = *projectDto.IsActive
	}
	return project
}

func (handler *syncHandler) fillInClientSideChangedTimeEntries(syncData *model.SyncData, changedTimeEntries []ChangedTimeEntryDto, userId uuid.UUID) {
	for _, changedTimeEntry := range changedTimeEntries {
		timeEntry, err := handler.createTimeEntryFromDto(changedTimeEntry, userId)
		if err == nil {
			switch changedTimeEntry.ChangeType {
			case NEW:
				syncData.TimeEntriesToBeCreated = append(syncData.TimeEntriesToBeCreated, timeEntry)
			case CHANGED:
				syncData.TimeEntriesToBeUpdated = append(syncData.TimeEntriesToBeUpdated, timeEntry)
			case DELETED:
				syncData.TimeEntriesToBeDeleted = append(syncData.TimeEntriesToBeDeleted, timeEntry)
			}
		} else {
			LogHandlerError("fillInClientSideChangedTimeEntries", err, "could not create time entry from DTO")
		}
	}
}

func (handler *syncHandler) createTimeEntryFromDto(timeEntryDto ChangedTimeEntryDto, userId uuid.UUID) (model.TimeEntry, error) {
	timeEntry := model.TimeEntry{
		ID:     timeEntryDto.Id,
		UserId: userId,
	}
	existingTimeEntry, err := handler.syncUsecase.GetTimeEntryById(timeEntryDto.Id)
	if err == nil && existingTimeEntry != nil {
		timeEntry = *existingTimeEntry
	}

	timeEntry.ProjectId = timeEntryDto.ProjectId

	startTime, err := time.Parse(time.RFC3339, timeEntryDto.StartTime)
	if err != nil {
		return model.TimeEntry{}, err
	}
	timeEntry.StartTime = startTime

	if timeEntryDto.Description != "" {
		timeEntry.Description = timeEntryDto.Description
	}
	if timeEntryDto.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, timeEntryDto.EndTime)
		if err != nil {
			return model.TimeEntry{}, err
		}
		timeEntry.EndTime = endTime
	}

	// Update deleted field if provided in DTO
	if timeEntryDto.Deleted != nil {
		timeEntry.Deleted = *timeEntryDto.Deleted
	}

	return timeEntry, nil
}
