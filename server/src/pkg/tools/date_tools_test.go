package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWeekNumber(t *testing.T) {
	assertWeekNumber(t, 2021, 9, 14, 37)
	assertWeekNumber(t, 2020, 12, 28, 53)
	assertWeekNumber(t, 2021, 1, 4, 1)
}

func assertWeekNumber(t *testing.T, year int, month time.Month, day int, expectedWeekNumber int) {
	weekNumber := GetWeekNumber(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, expectedWeekNumber, weekNumber)
}

func TestFirstDayOfFirstWeek(t *testing.T) {
	assertFirstDayOfFirstWeek(t, 2020, 2019, 12, 30)
	assertFirstDayOfFirstWeek(t, 2021, 2021, 1, 4)
	assertFirstDayOfFirstWeek(t, 2005, 2005, 1, 3)
	assertFirstDayOfFirstWeek(t, 2022, 2022, 1, 3)
	assertFirstDayOfFirstWeek(t, 2023, 2023, 1, 2)
	assertFirstDayOfFirstWeek(t, 2024, 2024, 1, 1)
	assertFirstDayOfFirstWeek(t, 2025, 2024, 12, 30)
}

func assertFirstDayOfFirstWeek(t *testing.T, year int, expectedYear int, expectedMonth time.Month, expectedDay int) {
	firstDayOfFirstWeek := GetFirstDayOfFirstWeek(year)
	assert.Equal(t, expectedYear, firstDayOfFirstWeek.Year())
	assert.Equal(t, expectedMonth, firstDayOfFirstWeek.Month())
	assert.Equal(t, expectedDay, firstDayOfFirstWeek.Day())
}

func TestFirstDayOfWeek(t *testing.T) {
	assertFirstDayOfWeek(t, 37, 2021, 2021, 9, 13)
	assertFirstDayOfWeek(t, 53, 2020, 2020, 12, 28)
	assertFirstDayOfWeek(t, 24, 2005, 2005, 6, 13)
	assertFirstDayOfWeek(t, 52, 2007, 2007, 12, 24)
	assertFirstDayOfWeek(t, 43, 2022, 2022, 10, 24)
	assertFirstDayOfWeek(t, 1, 2025, 2024, 12, 30)
}

func assertFirstDayOfWeek(t *testing.T, weekNumber int, year int, expectedYear int, expectedMonth time.Month, expectedDay int) {
	firstDayOfWeek := GetFirstDayOfWeek(weekNumber, year)
	assert.Equal(t, expectedYear, firstDayOfWeek.Year())
	assert.Equal(t, expectedMonth, firstDayOfWeek.Month())
	assert.Equal(t, expectedDay, firstDayOfWeek.Day())
}

func TestLastDayOfWeek(t *testing.T) {
	assertLastDayOfWeek(t, 37, 2021, 2021, 9, 19)
	assertLastDayOfWeek(t, 53, 2020, 2021, 1, 3)
	assertLastDayOfWeek(t, 24, 2005, 2005, 6, 19)
	assertLastDayOfWeek(t, 52, 2007, 2007, 12, 30)
}

func assertLastDayOfWeek(t *testing.T, weekNumber int, year int, expectedYear int, expectedMonth time.Month, expectedDay int) {
	lastDayOfWeek := GetLastDayOfWeek(weekNumber, year)
	assert.Equal(t, expectedYear, lastDayOfWeek.Year())
	assert.Equal(t, expectedMonth, lastDayOfWeek.Month())
	assert.Equal(t, expectedDay, lastDayOfWeek.Day())
}
