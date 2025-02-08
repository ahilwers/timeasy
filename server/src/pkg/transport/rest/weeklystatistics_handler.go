package rest

import (
	"github.com/gin-gonic/gin"
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

func NewWeeklyStatisticsHandler(tokenVerifier TokenVerifier, usecase *usecase.WeeklyStatisticsUsecase) WeeklyStatisticsHandler {
	return &weeklyStatisticsHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
	}
}

type weeklyStatisticsHandler struct {
	tokenVerifier TokenVerifier
	usecase       *usecase.WeeklyStatisticsUsecase
}

type weeklyStatisticsDto struct {
	WeekNumber int                  `json:"weekNumber"`
	Year       int                  `json:"year"`
	FirstDay   time.Time            `json:"firstDay"`
	LastDay    time.Time            `json:"lastDay"`
	Days       []dailyStatisticsDto `json:"days"`
}

type dailyStatisticsDto struct {
	Weekday       string `json:"weekday"`
	TimeInSeconds int    `json:"timeInSeconds"`
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

	dto := handler.createWeeklyStatisticsDto(*weeklyStatistics, weekNumber, year)
	context.JSON(http.StatusOK, dto)
}

func (handler *weeklyStatisticsHandler) createWeeklyStatisticsDto(ws model.WeeklyStatistics, weekNumber int, year int) weeklyStatisticsDto {
	dto := weeklyStatisticsDto{}
	dto.FirstDay = tools.GetFirstDayOfWeek(weekNumber, year)
	dto.LastDay = tools.GetLastDayOfWeek(weekNumber, year)
	days := make([]dailyStatisticsDto, 7)
	for i := 0; i < 7; i++ {
		var weekday time.Weekday
		if i == 6 {
			weekday = time.Sunday
		} else {
			weekday = time.Weekday(i + 1)
		}
		seconds := ws.GetSecondsForWeekday(weekday)
		days[i] = dailyStatisticsDto{
			Weekday:       handler.getWeekDayAsString(weekday),
			TimeInSeconds: seconds,
		}
	}
	dto.WeekNumber = weekNumber
	dto.Year = year
	dto.Days = days
	return dto
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
