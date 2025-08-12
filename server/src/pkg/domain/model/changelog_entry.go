package model

import (
	"github.com/gofrs/uuid"
	"time"
)

type EntityType string

const (
	EntityTypeProject            EntityType = "Project"
	EntityTypeTimeEntry          EntityType = "TimeEntry"
	EntityTypeTeam               EntityType = "Team"
	EntityTypeUserTeamAssignment EntityType = "UserTeamAssignment"
)

type Operation string

const (
	OperationCreated Operation = "Created"
	OperationUpdated Operation = "Updated"
	OperationDeleted Operation = "Deleted"
)

type ChangelogEntry struct {
	ID              int64
	EntityType      EntityType
	EntityID        uuid.UUID
	Operation       Operation
	ChangedByUser   uuid.UUID
	ChangedByClient string
	ChangedAt       time.Time
}
