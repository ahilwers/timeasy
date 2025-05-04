package usecase

import (
	"github.com/gofrs/uuid"
	"time"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/tools"
)

type WeeklyStatisticsUsecase struct {
	timeEntryUsecase TimeEntryUsecase
}

func NewWeeklyStatisticsUsecase(timeEntryUsecase TimeEntryUsecase) *WeeklyStatisticsUsecase {
	return &WeeklyStatisticsUsecase{
		timeEntryUsecase: timeEntryUsecase,
	}
}

// Creates weekly statistics for the specified user and project. if the projectId is nil, all projects will be taken into account
func (u *WeeklyStatisticsUsecase) Build(userId uuid.UUID, projectId uuid.UUID, weekNumber int, year int) (*model.WeeklyStatistics, error) {
	weeklyStatistics := model.NewWeeklyStatistics()
	startDate := tools.GetFirstDayOfWeek(weekNumber, year)
	endDate := tools.GetLastDayOfWeek(weekNumber, year)
	timeEntries, err := u.timeEntryUsecase.GetTimeEntriesOfUserAndProjectBetweenDates(userId, projectId, startDate, endDate)
	if err != nil {
		return nil, err
	}

	lastDay := 0
	for _, timeEntry := range timeEntries {
		if !u.isTimeEntryValid(timeEntry, endDate) {
			continue
		}
		currentDay := timeEntry.StartTime.Day()
		statisticsEntry := weeklyStatistics.GetEntryForWeekday(timeEntry.StartTime.Weekday())
		if (currentDay > lastDay) || (statisticsEntry == nil) {
			statisticsEntry = model.NewWeeklyStatisticsEntry(timeEntry.StartTime)
			weeklyStatistics.AddEntryForWeekday(timeEntry.StartTime.Weekday(), statisticsEntry)
		}
		statisticsEntry.Seconds += timeEntry.GetSeconds()
		statisticsEntry.AddSecondsForProject(timeEntry.ProjectId, timeEntry.GetSeconds())
		lastDay = currentDay
	}
	return weeklyStatistics, nil
}

func (u *WeeklyStatisticsUsecase) isTimeEntryValid(timeEntry model.TimeEntry, endDate time.Time) bool {
	nextDate := endDate.AddDate(0, 0, 1)
	return timeEntry.StartTime.Before(nextDate)
}
