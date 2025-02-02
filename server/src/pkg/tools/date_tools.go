package tools

import (
	"time"
)

func GetWeekNumber(date time.Time) int {
	dayOfYear := date.YearDay()
	weekNumber := (dayOfYear - int(date.Weekday()) + 10) / 7
	if weekNumber < 1 {
		weekNumber = GetNumberOfWeeks(date.Year() - 1)
	} else if weekNumber > GetNumberOfWeeks(date.Year()) {
		weekNumber = 1
	}
	return weekNumber
}

func GetNumberOfWeeks(year int) int {
	december28 := time.Date(year, time.December, 28, 0, 0, 0, 0, time.UTC)
	dayOfDecember28 := december28.YearDay()
	return (dayOfDecember28 - int(december28.Weekday()) + 10) / 7
}

func GetFirstDayOfWeek(weekNumber int, year int) time.Time {
	daysInYear := (weekNumber - 1) * 7
	firstDayOfFirstWeek := GetFirstDayOfFirstWeek(year)
	return firstDayOfFirstWeek.AddDate(0, 0, daysInYear)
}

func GetFirstDayOfFirstWeek(year int) time.Time {
	december28 := time.Date(year-1, time.December, 28, 0, 0, 0, 0, time.UTC)
	firstDay := december28.AddDate(0, 0, 7)
	if firstDay.Weekday() != time.Monday {
		firstDay = firstDay.AddDate(0, 0, -int(firstDay.Weekday())+1)
	}
	return firstDay
}

func GetLastDayOfWeek(weekNumber int, year int) time.Time {
	firstDayOfWeek := GetFirstDayOfWeek(weekNumber, year)
	return firstDayOfWeek.AddDate(0, 0, 6)
}

func OnlyDate(dateTime time.Time) time.Time {
	return time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, time.UTC)
}
