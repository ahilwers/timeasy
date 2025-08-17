package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"
)

type TimeEntryExportHandler interface {
	ExportTimeEntriesToXls(context *gin.Context)
	ExportTimeEntriesToCsv(context *gin.Context)
	ExportTimeEntriesToXlsOneLinePerDay(context *gin.Context)
}

type ExportFormat int

const (
	Csv ExportFormat = iota
	Xlsx
	XlsxOneLinePerDay
)

type timeEntryExportHandler struct {
	tokenVerifier    TokenVerifier
	timeEntryUsecase usecase.TimeEntryUsecase
	exportUsecase    usecase.TimeEntryExportUsecase
}

func NewTimeEntryExportHandler(tokenVerifier TokenVerifier, timeEntryUsecase usecase.TimeEntryUsecase) TimeEntryExportHandler {
	return &timeEntryExportHandler{
		tokenVerifier:    tokenVerifier,
		timeEntryUsecase: timeEntryUsecase,
	}
}

func (t timeEntryExportHandler) ExportTimeEntriesToXls(context *gin.Context) {
	t.exportTimeEntries(context, Xlsx)
}

func (t timeEntryExportHandler) ExportTimeEntriesToCsv(context *gin.Context) {
	t.exportTimeEntries(context, Csv)
}

func (t timeEntryExportHandler) ExportTimeEntriesToXlsOneLinePerDay(context *gin.Context) {
	t.exportTimeEntries(context, XlsxOneLinePerDay)

}

func (t timeEntryExportHandler) exportTimeEntries(context *gin.Context, format ExportFormat) {
	token, err := t.tokenVerifier.VerifyToken(context)
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "token verification failed")
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	userId, err := token.GetUserId()
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "failed to get user ID from token")
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
		startDate, err = time.Parse(time.DateOnly, startDateStr)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid startDate format, must be RFC3339"})
			return
		}
		endDate, err = time.Parse(time.DateOnly, endDateStr)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid endDate format, must be RFC3339"})
			return
		}
	}

	var timeEntries []model.TimeEntry
	if projectId == uuid.Nil && startDate.IsZero() && endDate.IsZero() {
		timeEntries, err = t.timeEntryUsecase.GetAllTimeEntriesOfUser(userId)
	} else if projectId != uuid.Nil && startDate.IsZero() && endDate.IsZero() {
		timeEntries, err = t.timeEntryUsecase.GetAllTimeEntriesOfUserAndProject(userId, projectId)
	} else {
		timeEntries, err = t.timeEntryUsecase.GetTimeEntriesOfUserAndProjectBetweenDates(userId, projectId, startDate, endDate)
	}
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "failed to get time entries")
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all entries"})
		return
	}

	exportUsecase, err := t.createExportUsecase(format)
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "failed to create export usecase")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	buffer, err := exportUsecase.ExportTimeEntries(timeEntries)
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "failed to export time entries")
		context.JSON(http.StatusInternalServerError, gin.H{"error": "error creating export"})
		return
	}
	contentType, err := t.getContentType(format)
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "failed to get content type")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fileName, err := t.getFileName(format)
	if err != nil {
		LogHandlerError("exportTimeEntries", err, "failed to get file name")
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.Header("Content-Description", "File Transfer")
	context.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	context.Data(http.StatusOK, contentType, buffer.Bytes())
}

func (t *timeEntryExportHandler) createExportUsecase(format ExportFormat) (usecase.TimeEntryExportUsecase, error) {
	switch format {
	case Csv:
		return usecase.NewTimeEntryExportToCsvUsecase(), nil
	case Xlsx:
		return usecase.NewTimeEntryExportToXlsxUsecase(), nil
	case XlsxOneLinePerDay:
		return usecase.NewTimeEntryExportToXlsxUsecaseOneLinePerDay(), nil
	}
	return nil, fmt.Errorf("could not create export usecase: unknown format")
}

func (t *timeEntryExportHandler) getContentType(format ExportFormat) (string, error) {
	switch format {
	case Csv:
		return "text/csv", nil
	case Xlsx:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	case XlsxOneLinePerDay:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	}
	return "", fmt.Errorf("could not determine content type: unknown format")
}

func (t *timeEntryExportHandler) getFileName(format ExportFormat) (string, error) {
	switch format {
	case Csv:
		return "time_entries.csv", nil
	case Xlsx:
		return "time_entries.xlsx", nil
	case XlsxOneLinePerDay:
		return "time_entries.xlsx", nil
	}
	return "", fmt.Errorf("could not determine file name: unknown format")
}
