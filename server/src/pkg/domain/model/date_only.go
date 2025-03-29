package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateOnly is a custom type for handling dates without time component.
type DateOnly time.Time

func NewDateOnly(t time.Time) DateOnly {
	// Strip the time part, keep only the date
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return DateOnly(date)
}

// ToTime converts DateOnly back to time.Time
func (d DateOnly) ToTime() time.Time {
	return time.Time(d)
}

func (d DateOnly) IsZero() bool {
	return time.Time(d).IsZero()
}

// MarshalJSON formats the date as "YYYY-MM-DD"
func (d DateOnly) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02"))
	return []byte(formatted), nil
}

// UnmarshalJSON parses a date in "YYYY-MM-DD" format
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	if str == "null" || str == "" {
		*d = DateOnly(time.Time{})
		return nil
	}
	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}
	*d = DateOnly(t)
	return nil
}

// Value returns the database driver compatible value
func (d DateOnly) Value() (driver.Value, error) {
	return time.Time(d), nil
}

// Scan reads the database value into DateOnly
func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		*d = DateOnly(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*d = DateOnly(v)
		return nil
	default:
		return fmt.Errorf("cannot scan value %v into DateOnly", value)
	}
}
