package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"net/http"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/tools"
	"timeasy-server/pkg/usecase"
)

type WeeklyStatisticsHandler interface {
	GetWeeklyStatistics(context *gin.Context)
	GetCurrentWeekNumber(context *gin.Context)
}

func NewWeeklyStatisticsHandler(tokenVerifier TokenVerifier, usecase *usecase.WeeklyStatisticsUsecase, projectUsecase usecase.ProjectUsecase) WeeklyStatisticsHandler {
	return &weeklyStatisticsHandler{
		tokenVerifier:  tokenVerifier,
		usecase:        usecase,
		projectUsecase: projectUsecase,
	}
}

type weeklyStatisticsHandler struct {
	tokenVerifier  TokenVerifier
	usecase        *usecase.WeeklyStatisticsUsecase
	projectUsecase usecase.ProjectUsecase
}

type weeklyStatisticsDto struct {
	WeekNumber   int                  `json:"weekNumber"`
	Year         int                  `json:"year"`
	FirstDay     string               `json:"firstDay"`
	LastDay      string               `json:"lastDay"`
	Days         []dailyStatisticsDto `json:"days"`
	SumInSeconds int                  `json:"sumInSeconds"`
}

type dailyStatisticsDto struct {
	Weekday         string               `json:"weekday"`
	TimeInSeconds   int                  `json:"timeInSeconds"`
	TimesPerProject *[]timePerProjectDto `json:"timesPerProject,omitempty"`
}

type timePerProjectDto struct {
	ProjectId     uuid.UUID `json:"projectId"`
	ProjectName   string    `json:"projectName"`
	ProjectColor  string    `json:"projectColor"`
	TimeInSeconds int       `json:"timeInSeconds"`
}

func (handler *weeklyStatisticsHandler) GetWeeklyStatistics(context *gin.Context) {
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

	weekNumber, err := GetMandatoryIntParamValue(context, "week")
	if err != nil || weekNumber < 1 || weekNumber > 53 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "please specify a valid week number"})
		return
	}

	year, err := GetMandatoryIntParamValue(context, "year")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "please specify a valid year"})
		return
	}

	projectId, err := GetOptionalIdParamValue(context, "project")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	weeklyStatistics, err := handler.usecase.Build(userId, projectId, weekNumber, year)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dto, err := handler.createWeeklyStatisticsDto(*weeklyStatistics, weekNumber, year, projectId)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, dto)
}

func (handler *weeklyStatisticsHandler) createWeeklyStatisticsDto(ws model.WeeklyStatistics, weekNumber int, year int, projectId uuid.UUID) (weeklyStatisticsDto, error) {
	dto := weeklyStatisticsDto{}
	dto.FirstDay = tools.GetFirstDayOfWeek(weekNumber, year).Format(time.RFC3339)
	dto.LastDay = tools.GetLastDayOfWeek(weekNumber, year).Format(time.RFC3339)
	days := make([]dailyStatisticsDto, 7)
	for i := 0; i < 7; i++ {
		var weekday time.Weekday
		if i == 6 {
			weekday = time.Sunday
		} else {
			weekday = time.Weekday(i + 1)
		}
		seconds := ws.GetSecondsForWeekday(weekday)
		dayEntry := ws.GetEntryForWeekday(weekday)
		timesPerProject := make([]timePerProjectDto, 0)
		if dayEntry != nil && projectId == uuid.Nil {
			for _, timePerProject := range dayEntry.TimesPerProject {
				timePerProjectDto, err := handler.createDtoFromTimePerProject(*timePerProject)
				if err != nil {
					return weeklyStatisticsDto{}, err
				}
				timesPerProject = append(timesPerProject, timePerProjectDto)
			}
		}
		days[i] = dailyStatisticsDto{
			Weekday:       handler.getWeekDayAsString(weekday),
			TimeInSeconds: seconds,
		}
		if projectId == uuid.Nil {
			days[i].TimesPerProject = &timesPerProject
		}
	}
	dto.WeekNumber = weekNumber
	dto.Year = year
	dto.Days = days
	dto.SumInSeconds = ws.GetSumInSeconds()
	return dto, nil
}

func (handler *weeklyStatisticsHandler) createDtoFromTimePerProject(timePerProject model.TimePerProject) (timePerProjectDto, error) {
	project, err := handler.projectUsecase.GetProjectById(timePerProject.ProjectId)
	if err != nil {
		return timePerProjectDto{}, err
	}
	return timePerProjectDto{
		ProjectId:     timePerProject.ProjectId,
		ProjectName:   project.Name,
		ProjectColor:  project.Color,
		TimeInSeconds: timePerProject.Seconds,
	}, nil
}

func (handler *weeklyStatisticsHandler) getWeekDayAsString(weekday time.Weekday) string {
	switch weekday {
	case time.Sunday:
		return "Sunday"
	case time.Monday:
		return "Monday"
	case time.Tuesday:
		return "Tuesday"
	case time.Wednesday:
		return "Wednesday"
	case time.Thursday:
		return "Thursday"
	case time.Friday:
		return "Friday"
	case time.Saturday:
		return "Saturday"
	default:
		return "Unknown"
	}
}

type currentWeekDto struct {
	WeekNumber int `json:"weekNumber"`
	Year       int `json:"year"`
}

func (handler *weeklyStatisticsHandler) GetCurrentWeekNumber(context *gin.Context) {
	now := time.Now()
	year, weekNumber := now.ISOWeek()

	dto := currentWeekDto{
		WeekNumber: weekNumber,
		Year:       year,
	}
	context.JSON(http.StatusOK, dto)
}
