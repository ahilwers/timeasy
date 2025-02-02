package model

import "time"

func NewWeeklyStatistics() *WeeklyStatistics {
	return &WeeklyStatistics{
		entries: make(map[time.Weekday]*WeeklyStatisticsEntry),
	}
}

type WeeklyStatistics struct {
	entries map[time.Weekday]*WeeklyStatisticsEntry
}

func (stats *WeeklyStatistics) AddEntryForWeekDay(weekDay time.Weekday, entry *WeeklyStatisticsEntry) {
	stats.entries[weekDay] = entry
}

func (stats *WeeklyStatistics) GetEntryForWeekDay(weekDay time.Weekday) *WeeklyStatisticsEntry {
	return stats.entries[weekDay]
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
	Date    time.Time
	Seconds int
}

func (entry *WeeklyStatisticsEntry) GetMinutes() float32 {
	return float32(entry.Seconds) / 60
}

func (entry *WeeklyStatisticsEntry) GetHours() float32 {
	return entry.GetMinutes() / 60
}
