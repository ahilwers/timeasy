package rest

import (
	"encoding/json"
	"fmt"
	"strings"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type ChangeType uint8

const (
	NEW ChangeType = iota
	CHANGED
	DELETED
)

var (
	ChangeType_Name = map[uint8]string{
		0: "NEW",
		1: "CHANGED",
		2: "DELETED",
	}

	ChangeType_Value = map[string]uint8{
		"NEW":     0,
		"CHANGED": 1,
		"DELETED": 2,
	}
)

func (c ChangeType) String() string {
	return ChangeType_Name[uint8(c)]
}

func (c ChangeType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *ChangeType) UnmarshalJSON(data []byte) (err error) {
	var sType string
	if err := json.Unmarshal(data, &sType); err != nil {
		return err
	}
	if *c, err = c.parse(sType); err != nil {
		return err
	}
	return nil
}

func (c *ChangeType) parse(sType string) (ChangeType, error) {
	sType = strings.TrimSpace(strings.ToUpper(sType))
	value, ok := ChangeType_Value[sType]
	if !ok {
		return ChangeType(0), fmt.Errorf("%v is not a valid change type", sType)
	}
	return ChangeType(value), nil
}

type SyncEntries struct {
	TimeEntries       []ChangedTimeEntryDto
	Projects          []ChangedProjectDto
	LatestChangeLogId int64
}

type ChangedTimeEntryDto struct {
	Id              uuid.UUID  `json:"id" binding:"required"`
	Description     string     `json:"description"`
	StartTime       string     `json:"startTime" binding:"required"`
	EndTime         string     `json:"endTime,omitempty"`
	ProjectId       uuid.UUID  `json:"projectId" binding:"required"`
	ChangeType      ChangeType `json:"changeType" binding:"required"`
	ChangeTimestamp string     `json:"changeTimestamp"`
}

type ChangedProjectDto struct {
	Id              uuid.UUID        `json:"id" binding:"required"`
	Name            string           `json:"name" binding:"required"`
	Color           *string          `json:"color"`
	Deadline        *model.DateOnly  `json:"deadline,omitempty"`
	HourlyRate      *float64         `json:"hourlyRate,omitempty"`
	TimeBudget      *int             `json:"timeBudget,omitempty"`
	IsActive        *bool            `json:"isActive,omitempty"`
	ChangeType      ChangeType       `json:"changeType" binding:"required"`
	ChangeTimestamp string           `json:"changeTimestamp"`
}
