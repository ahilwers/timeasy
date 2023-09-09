package rest

import (
	"encoding/json"
	"fmt"
	"strings"

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

type ChangedTimeEntryDto struct {
	Id               uuid.UUID
	Description      string `json:"description" binding:"required"`
	StartTimeUTCUnix int64  `json:"startTimeUTCUnix" binding:"required"`
	EndTimeUTCUnix   int64
	ProjectId        uuid.UUID  `json:"projectId" binding:"required"`
	ChangeType       ChangeType `json:"changeType" binding:"required"`
}
