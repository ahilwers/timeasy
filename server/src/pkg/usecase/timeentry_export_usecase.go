package usecase

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/xuri/excelize/v2"
	"sort"
	"time"
	"timeasy-server/pkg/domain/model"
)

type TimeEntryExportUsecase interface {
	ExportTimeEntries(timeEntries []model.TimeEntry) (*bytes.Buffer, error)
}

type timeEntryExportToCsvUsecase struct {
}

func NewTimeEntryExportToCsvUsecase() TimeEntryExportUsecase {
	return &timeEntryExportToCsvUsecase{}
}

type timeEntryExportToXlsxUsecase struct {
}

func NewTimeEntryExportToXlsxUsecase() TimeEntryExportUsecase {
	return &timeEntryExportToXlsxUsecase{}
}

type timeEntryExportToXlsxUsecaseOneLinePerDay struct {
}

func NewTimeEntryExportToXlsxUsecaseOneLinePerDay() TimeEntryExportUsecase {
	return &timeEntryExportToXlsxUsecaseOneLinePerDay{}
}

func (t *timeEntryExportToCsvUsecase) ExportTimeEntries(timeEntries []model.TimeEntry) (*bytes.Buffer, error) {
	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)

	writer.Comma = ';'

	headers := []string{"Start Time", "End Time", "Description"}
	if err := writer.Write(headers); err != nil {
		return nil, fmt.Errorf("error writing header to csv: %w", err)
	}

	for _, entry := range timeEntries {
		endTime := ""
		if !entry.EndTime.IsZero() {
			endTime = entry.EndTime.Format(time.RFC3339)
		}
		record := []string{
			fmt.Sprintf(`"%s"`, entry.StartTime.Format(time.RFC3339)),
			fmt.Sprintf(`"%s"`, endTime),
			fmt.Sprintf(`"%s"`, entry.Description),
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("error writing record to csv: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing csv writer: %w", err)
	}

	return &csvBuffer, nil
}

func (t timeEntryExportToXlsxUsecase) ExportTimeEntries(timeEntries []model.TimeEntry) (*bytes.Buffer, error) {
	excelFile := excelize.NewFile()
	defer func() {
		if err := excelFile.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	sheetName := "Sheet1"
	headers := []string{"Start Date", "Start Time", "End Time", "Description"}
	for col, header := range headers {
		colName, _ := excelize.ColumnNumberToName(col + 1)
		cell := fmt.Sprintf("%s1", colName)
		excelFile.SetCellValue(sheetName, cell, header)
	}
	loc, _ := time.LoadLocation("Local")
	dateStyle, _ := excelFile.NewStyle(&excelize.Style{NumFmt: 14}) // Standard-Date (YYYY-MM-DD)
	timeStyle, _ := excelFile.NewStyle(&excelize.Style{NumFmt: 20}) // Standard-Time (HH:MM)
	for i, entry := range timeEntries {
		row := i + 2 // Daten beginnen ab Zeile 2
		startTime := entry.StartTime.In(loc)

		var endTime interface{} = ""
		if !entry.EndTime.IsZero() {
			endTime = entry.EndTime.In(loc)
		}

		excelFile.SetCellValue(sheetName, fmt.Sprintf("A%d", row), startTime)
		excelFile.SetCellValue(sheetName, fmt.Sprintf("B%d", row), startTime)
		excelFile.SetCellValue(sheetName, fmt.Sprintf("C%d", row), endTime)
		excelFile.SetCellValue(sheetName, fmt.Sprintf("D%d", row), entry.Description)

		excelFile.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("E%d", row), dateStyle)
		excelFile.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("F%d", row), timeStyle)
		excelFile.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("H%d", row), timeStyle)
	}

	excelFile.SetColWidth(sheetName, "A", "I", 15)

	excelFile.SetActiveSheet(0)

	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (t timeEntryExportToXlsxUsecaseOneLinePerDay) ExportTimeEntries(timeEntries []model.TimeEntry) (*bytes.Buffer, error) {
	excelFile := excelize.NewFile()
	defer func() {
		if err := excelFile.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	sheetName := "Sheet1"
	t.WriteHeader(sheetName, excelFile)
	t.sortTimeEntries(timeEntries)
	dayTimeEntries := []model.TimeEntry{}
	lastDay := time.Time{}
	currentRow := 1
	for _, entry := range timeEntries {
		var entryDate = t.OnlyDate(entry.StartTime)
		if lastDay.IsZero() || entryDate != lastDay {
			t.WriteEntriesToFile(dayTimeEntries, sheetName, currentRow, excelFile)
			dayTimeEntries = []model.TimeEntry{}
			lastDay = entryDate
			currentRow++
		}
		dayTimeEntries = append(dayTimeEntries, entry)
	}
	err := t.WriteEntriesToFile(dayTimeEntries, sheetName, currentRow, excelFile)
	if err != nil {
		return nil, err
	}
	excelFile.SetColWidth(sheetName, "D", "M", 14)
	excelFile.SetActiveSheet(0)

	buffer, err := excelFile.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (t timeEntryExportToXlsxUsecaseOneLinePerDay) WriteHeader(sheetName string, excelFile *excelize.File) error {
	excelFile.SetCellValue(sheetName, "A1", "Date")
	excelFile.SetCellValue(sheetName, "B1", "Start")
	excelFile.SetCellValue(sheetName, "C1", "End")
	pauseCellIndex := 3
	for i := 0; i < 5; i++ {
		pauseCellIndex++
		colName, _ := excelize.ColumnNumberToName(pauseCellIndex)
		excelFile.SetCellValue(sheetName, fmt.Sprintf("%s1", colName), fmt.Sprintf("Pause %d Start", i+1))
		pauseCellIndex++
		colName, _ = excelize.ColumnNumberToName(pauseCellIndex)
		excelFile.SetCellValue(sheetName, fmt.Sprintf("%s1", colName), fmt.Sprintf("Pause %d End", i+1))
	}
	return nil
}

func (t timeEntryExportToXlsxUsecaseOneLinePerDay) WriteEntriesToFile(timeEntries []model.TimeEntry, sheetName string, row int, excelFile *excelize.File) error {
	if len(timeEntries) == 0 {
		return nil
	}
	loc, _ := time.LoadLocation("Local")
	dateStyle, _ := excelFile.NewStyle(&excelize.Style{NumFmt: 14}) // Standard-Date (YYYY-MM-DD)
	timeStyle, _ := excelFile.NewStyle(&excelize.Style{NumFmt: 20}) // Standard-Time (HH:MM)

	startTime := timeEntries[0].StartTime.In(loc)
	endTime := timeEntries[len(timeEntries)-1].EndTime.In(loc)
	excelFile.SetCellValue(sheetName, fmt.Sprintf("A%d", row), startTime)
	excelFile.SetCellValue(sheetName, fmt.Sprintf("B%d", row), startTime)
	excelFile.SetCellValue(sheetName, fmt.Sprintf("C%d", row), endTime)

	colIndex := 3
	for i := 0; i < len(timeEntries)-1; i++ {
		entry := timeEntries[i]
		if entry.EndTime.IsZero() {
			continue
		}
		pauseStartTime := entry.EndTime.In(loc)

		nextEntry := timeEntries[i+1]
		pauseEndTime := nextEntry.StartTime.In(loc)

		colIndex++
		col, err := excelize.ColumnNumberToName(colIndex)
		if err != nil {
			return err
		}
		excelFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", col, row), pauseStartTime)
		colIndex++
		col, err = excelize.ColumnNumberToName(colIndex)
		if err != nil {
			return err
		}
		excelFile.SetCellValue(sheetName, fmt.Sprintf("%s%d", col, row), pauseEndTime)

	}
	excelFile.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dateStyle)
	excelFile.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("Z%d", row), timeStyle)
	return nil
}

func (t timeEntryExportToXlsxUsecaseOneLinePerDay) OnlyDate(fullTime time.Time) time.Time {
	return fullTime.Truncate(24 * time.Hour)
}

func (t timeEntryExportToXlsxUsecaseOneLinePerDay) sortTimeEntries(timeEntries []model.TimeEntry) {
	sort.Slice(timeEntries, func(i, j int) bool {
		if timeEntries[i].StartTime.Before(timeEntries[j].StartTime) {
			return true
		}
		if timeEntries[i].StartTime.After(timeEntries[j].StartTime) {
			return false
		}

		iHasEndTime := !timeEntries[i].EndTime.IsZero()
		jHasEndTime := !timeEntries[j].EndTime.IsZero()

		if iHasEndTime && jHasEndTime {
			return timeEntries[i].EndTime.Before(timeEntries[j].EndTime)
		}
		if !iHasEndTime && jHasEndTime {
			return false
		}
		if iHasEndTime && !jHasEndTime {
			return true
		}
		return false
	})
}
