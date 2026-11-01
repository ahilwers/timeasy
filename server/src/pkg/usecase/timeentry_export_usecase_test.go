package usecase

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"timeasy-server/pkg/domain/model"
)

func TestCsvExportDoesNotDoubleQuoteFields(t *testing.T) {
	usecase := NewTimeEntryExportToCsvUsecase()

	start := time.Date(2025, time.March, 24, 10, 30, 0, 0, time.UTC)
	end := time.Date(2025, time.March, 24, 11, 45, 0, 0, time.UTC)
	entries := []model.TimeEntry{
		{StartTime: start, EndTime: end, Description: "A plain description"},
		{StartTime: start.Add(time.Hour)},
		{StartTime: start.Add(2 * time.Hour), EndTime: end.Add(2 * time.Hour), Description: `Description with "quotes" and ; semicolon`},
	}

	buffer, err := usecase.ExportTimeEntries(entries)
	if err != nil {
		t.Fatalf("ExportTimeEntries returned an error: %v", err)
	}

	content := buffer.String()

	// Timestamps and plain descriptions must not be wrapped in redundant quotes.
	if strings.Contains(content, `"""`) {
		t.Errorf("CSV output contains triple quotes:\n%s", content)
	}
	if strings.Contains(content, `"2025-03-24T10:30:00Z"`) {
		t.Errorf("plain timestamp was wrapped in quotes:\n%s", content)
	}
	if strings.Contains(content, `"A plain description"`) {
		t.Errorf("plain description was wrapped in quotes:\n%s", content)
	}

	// Fields containing the delimiter or a quote character must be quoted with
	// the quote character escaped exactly once (standard CSV escaping).
	if !strings.Contains(content, `"Description with ""quotes"" and ; semicolon"`) {
		t.Errorf("description with special characters was not escaped correctly:\n%s", content)
	}

	// The output must parse cleanly as a semicolon-delimited CSV file.
	reader := csv.NewReader(bytes.NewReader(buffer.Bytes()))
	reader.Comma = ';'
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse generated CSV: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("expected 4 records (header + 3 entries), got %d", len(records))
	}

	wantHeader := []string{"Start Time", "End Time", "Description"}
	for i, value := range wantHeader {
		if records[0][i] != value {
			t.Errorf("header column %d: expected %q, got %q", i, value, records[0][i])
		}
	}
	if records[1][0] != "2025-03-24T10:30:00Z" || records[1][1] != "2025-03-24T11:45:00Z" || records[1][2] != "A plain description" {
		t.Errorf("first data record is malformed: %v", records[1])
	}
	if records[2][1] != "" {
		t.Errorf("missing end time should stay empty, got %q", records[2][1])
	}
	if records[3][2] != `Description with "quotes" and ; semicolon` {
		t.Errorf("quoted description was not parsed back correctly: %q", records[3][2])
	}
}
