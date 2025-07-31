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
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	syncEntries.LatestChangeLogId = latestChangeLogEntry

	entries, err := handler.syncUsecase.GetChangedTimeEntries(userId, sinceChangeLogEntry, latestChangeLogEntry, clientId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	handler.appendChangedTimeEntries(entries.Created, syncEntries, NEW)
	handler.appendChangedTimeEntries(entries.Updated, syncEntries, CHANGED)
	handler.appendChangedTimeEntries(entries.Deleted, syncEntries, DELETED)

	projects, err := handler.syncUsecase.GetChangedProjects(userId, sinceChangeLogEntry, latestChangeLogEntry, clientId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	handler.appendChangedProjects(projects.Created, syncEntries, NEW)
	handler.appendChangedProjects(projects.Updated, syncEntries, CHANGED)
	handler.appendChangedProjects(projects.Deleted, syncEntries, DELETED)

	context.JSON(http.StatusOK, syncEntries)
}

func (handler *syncHandler) appendChangedTimeEntries(timeEntries []model.TimeEntry, syncEntries SyncEntries, changeType ChangeType) {
	for _, entry := range timeEntries {
		desc := entry.Description
		syncTimeEntry := ChangedTimeEntryDto{
			Id:          entry.ID,
			Description: &desc,
			StartTime:   entry.StartTime.Format(time.RFC3339),
			ProjectId:   entry.ProjectId,
			ChangeType:  changeType,
		}
		if !entry.EndTime.IsZero() {
			syncTimeEntry.EndTime = entry.EndTime.Format(time.RFC3339)
		}
		syncEntries.TimeEntries = append(syncEntries.TimeEntries, syncTimeEntry)
	}
}

func (handler *syncHandler) appendChangedProjects(projects []model.Project, syncEntries SyncEntries, changeType ChangeType) {
	for _, project := range projects {
		deadline := project.Deadline
		hourlyRate := project.HourlyRate
		timeBudget := project.TimeBudget
		isActive := project.IsActive
		color := project.Color
		syncProject := ChangedProjectDto{
			Id:         project.ID,
			Name:       project.Name,
			Color:      &color,
			Deadline:   &deadline,
			HourlyRate: &hourlyRate,
			TimeBudget: &timeBudget,
			IsActive:   &isActive,
			ChangeType: changeType,
		}
		syncEntries.Projects = append(syncEntries.Projects, syncProject)
	}
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
	clientId := context.Query("clientId")

	var syncData model.SyncData
	handler.fillInClientSideChangedProjects(&syncData, syncDtos.Projects, userId)
	handler.fillInClientSideChangedTimeEntries(&syncData, syncDtos.TimeEntries, userId)

	err = handler.syncUsecase.UpdateAndDeleteData(syncData, userId, clientId)
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
		project.HourlyRate = *projectDto.HourlyRate
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

	if timeEntryDto.Description != nil {
		timeEntry.Description = *timeEntryDto.Description
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
