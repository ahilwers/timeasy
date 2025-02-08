package usecase

import (
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"timeasy-server/pkg/domain/model"
)

func Test_WeeklyStatisticsUseCase_CalculateWeeklyStatistics(t *testing.T) {
	useCaseTest := NewUsecaseTest()
	teardownTest := useCaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, useCaseTest.ProjectUsecase, "project", userId)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)
	// Thursday
	startTime = time.Date(2025, time.January, 2, 11, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 2, 11, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)
	// Sunday
	startTime = time.Date(2025, time.January, 5, 12, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 5, 12, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)

	weeklyStatisticsUsecase := NewWeeklyStatisticsUsecase(useCaseTest.TimeEntryUsecase)
	weeklyStatistics, err := weeklyStatisticsUsecase.Build(userId, project.ID, 1, 2025)
	assert.Nil(t, err)
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Monday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Tuesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Wednesday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Thursday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Friday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Saturday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Sunday))
}

func Test_WeeklyStatisticsUseCase_CalculatingWeeklyStatisticsIgnoresEntriesFromOtherWeek(t *testing.T) {
	useCaseTest := NewUsecaseTest()
	teardownTest := useCaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	project := addProject(t, useCaseTest.ProjectUsecase, "project", userId)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)
	// Monday following Week
	startTime = time.Date(2025, time.January, 6, 11, 0, 0, 0, time.UTC)
	endTime = time.Date(2025, time.January, 6, 11, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, project.ID, startTime, endTime)

	weeklyStatisticsUsecase := NewWeeklyStatisticsUsecase(useCaseTest.TimeEntryUsecase)
	weeklyStatistics, err := weeklyStatisticsUsecase.Build(userId, project.ID, 1, 2025)
	assert.Nil(t, err)
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Monday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Tuesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Wednesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Thursday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Friday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Saturday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Sunday))
}

func Test_WeeklyStatisticsUseCase_CalculatingWeeklyStatisticsIgnoresEntriesFromOtherUser(t *testing.T) {
	useCaseTest := NewUsecaseTest()
	teardownTest := useCaseTest.SetupTest(t)
	defer teardownTest(t)

	firstUserId := GetTestUserId(t)
	secondUserId := GetTestUserId(t)
	project := addProject(t, useCaseTest.ProjectUsecase, "project", firstUserId)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, firstUserId, project.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, firstUserId, project.ID, startTime, endTime)
	// Monday times of other user
	startTime = time.Date(2024, time.December, 30, 16, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 30, 16, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, secondUserId, project.ID, startTime, endTime)

	weeklyStatisticsUsecase := NewWeeklyStatisticsUsecase(useCaseTest.TimeEntryUsecase)
	weeklyStatistics, err := weeklyStatisticsUsecase.Build(firstUserId, project.ID, 1, 2025)
	assert.Nil(t, err)
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Monday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Tuesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Wednesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Thursday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Friday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Saturday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Sunday))
}

func Test_WeeklyStatisticsUseCase_CalculatingWeeklyStatisticsIgnoresEntriesFromOtherProject(t *testing.T) {
	useCaseTest := NewUsecaseTest()
	teardownTest := useCaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	firstProject := addProject(t, useCaseTest.ProjectUsecase, "firstProject", userId)
	secondProject := addProject(t, useCaseTest.ProjectUsecase, "secondProject", userId)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, firstProject.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, firstProject.ID, startTime, endTime)
	// Monday times from other firstProject
	startTime = time.Date(2024, time.December, 30, 16, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 30, 16, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, secondProject.ID, startTime, endTime)

	weeklyStatisticsUsecase := NewWeeklyStatisticsUsecase(useCaseTest.TimeEntryUsecase)
	weeklyStatistics, err := weeklyStatisticsUsecase.Build(userId, firstProject.ID, 1, 2025)
	assert.Nil(t, err)
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Monday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Tuesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Wednesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Thursday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Friday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Saturday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Sunday))
}

func Test_WeeklyStatisticsUseCase_CalculatingWeeklyStatisticsUsesAllProjectsIfProjectIsNil(t *testing.T) {
	useCaseTest := NewUsecaseTest()
	teardownTest := useCaseTest.SetupTest(t)
	defer teardownTest(t)

	userId := GetTestUserId(t)
	firstProject := addProject(t, useCaseTest.ProjectUsecase, "firstProject", userId)
	secondProject := addProject(t, useCaseTest.ProjectUsecase, "secondProject", userId)

	// Monday
	startTime := time.Date(2024, time.December, 30, 15, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, time.December, 30, 15, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, firstProject.ID, startTime, endTime)
	// Tuesday
	startTime = time.Date(2024, time.December, 31, 10, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 31, 10, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, firstProject.ID, startTime, endTime)
	// Monday times from other firstProject
	startTime = time.Date(2024, time.December, 30, 16, 0, 0, 0, time.UTC)
	endTime = time.Date(2024, time.December, 30, 16, 30, 0, 0, time.UTC)
	AddTimeEntry(t, useCaseTest, userId, secondProject.ID, startTime, endTime)

	weeklyStatisticsUsecase := NewWeeklyStatisticsUsecase(useCaseTest.TimeEntryUsecase)
	weeklyStatistics, err := weeklyStatisticsUsecase.Build(userId, uuid.Nil, 1, 2025)
	assert.Nil(t, err)
	assert.Equal(t, 3600, weeklyStatistics.GetSecondsForWeekday(time.Monday))
	assert.Equal(t, 1800, weeklyStatistics.GetSecondsForWeekday(time.Tuesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Wednesday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Thursday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Friday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Saturday))
	assert.Equal(t, 0, weeklyStatistics.GetSecondsForWeekday(time.Sunday))
}

func AddTimeEntry(t *testing.T, usecaseTest *UsecaseTest, userId uuid.UUID, projectId uuid.UUID, startTime time.Time, endTime time.Time) *model.TimeEntry {
	timeEntry := model.TimeEntry{
		Description: "timeentry",
		StartTime:   startTime,
		EndTime:     endTime,
		UserId:      userId,
		ProjectId:   projectId,
	}
	err := usecaseTest.TimeEntryUsecase.AddTimeEntry(&timeEntry)
	assert.Nil(t, err)
	return &timeEntry
}
