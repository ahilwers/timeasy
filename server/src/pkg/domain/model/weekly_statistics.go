package model

import (
	"github.com/gofrs/uuid"
	"time"
)

func NewWeeklyStatistics() *WeeklyStatistics {
	return &WeeklyStatistics{
		entries: make(map[time.Weekday]*WeeklyStatisticsEntry),
	}
}

type WeeklyStatistics struct {
	entries map[time.Weekday]*WeeklyStatisticsEntry
}

func (stats *WeeklyStatistics) AddEntryForWeekday(weekday time.Weekday, entry *WeeklyStatisticsEntry) {
	stats.entries[weekday] = entry
}

func (stats *WeeklyStatistics) GetEntryForWeekday(weekday time.Weekday) *WeeklyStatisticsEntry {
	return stats.entries[weekday]
}

func (stats *WeeklyStatistics) GetSecondsForWeekday(weekday time.Weekday) int {
	entry := stats.GetEntryForWeekday(weekday)
	if entry == nil {
		return 0
	}
	return entry.Seconds
}

func (stats *WeeklyStatistics) GetSumInSeconds() int {
	var sum int
	for _, entry := range stats.entries {
		sum += entry.Seconds
	}
	return sum
}

func NewWeeklyStatisticsEntry(date time.Time) *WeeklyStatisticsEntry {
	return &WeeklyStatisticsEntry{
		Date: date,
	}
}

type WeeklyStatisticsEntry struct {
	Date            time.Time
	Seconds         int
	TimesPerProject map[uuid.UUID]*TimePerProject
}

type TimePerProject struct {
	ProjectId uuid.UUID
	Seconds   int
}

func (entry *WeeklyStatisticsEntry) GetMinutes() float32 {
	return float32(entry.Seconds) / 60
}

func (entry *WeeklyStatisticsEntry) GetHours() float32 {
	return entry.GetMinutes() / 60
}

func (entry *WeeklyStatisticsEntry) AddSecondsForProject(projectId uuid.UUID, seconds int) {
	if entry.TimesPerProject == nil {
		entry.TimesPerProject = make(map[uuid.UUID]*TimePerProject)
	}
	timeOfProject := entry.TimesPerProject[projectId]
	if timeOfProject == nil {
		timeOfProject = &TimePerProject{
			ProjectId: projectId,
			Seconds:   0,
		}
	}
	timeOfProject.Seconds += seconds
	entry.TimesPerProject[projectId] = timeOfProject
}
