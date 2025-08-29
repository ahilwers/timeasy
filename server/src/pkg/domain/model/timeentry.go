package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type TimeEntry struct {
	ID                  uuid.UUID
	UserId              uuid.UUID
	ProjectId           uuid.UUID
	Project             Project
	StartTime           time.Time
	EndTime             time.Time
	Description         string
	ExternalIssueID     *uuid.UUID `json:"externalIssueId,omitempty"`
	PendingExternalRef  *string `json:"pendingExternalRef,omitempty"`
	Deleted             bool
}

func (timeEntry *TimeEntry) GetSeconds() int {
	var myEndTime time.Time
	if timeEntry.EndTime.IsZero() {
		myEndTime = time.Now()
	} else {
		myEndTime = timeEntry.EndTime
	}
	return int(myEndTime.Sub(timeEntry.StartTime).Seconds())
}
