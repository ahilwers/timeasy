package repository

import (
	"github.com/gofrs/uuid"
	"time"

	"timeasy-server/pkg/domain/model"
)

// ChangelogFilter contains filter criteria for querying changelog entries
type ChangelogFilter struct {
	// EntityType filters by the type of entity (e.g., Project, TimeEntry, etc.)
	EntityType model.EntityType

	// EntityID filters by the ID of the entity
	EntityID uuid.UUID

	// ChangedByUser filters by the user who made the change
	ChangedByUser uuid.UUID

	// Operation filters by the type of change (Created, Updated, Deleted)
	Operation model.Operation

	// StartDate filters entries changed on or after this date
	StartDate time.Time

	// EndDate filters entries changed on or before this date
	EndDate time.Time

	// Limit limits the number of results returned
	Limit int
}

type ChangelogRepository interface {
	AddChangelogEntry(entry *model.ChangelogEntry, tx model.Transaction) error
	GetChangelogEntries(filter *ChangelogFilter) ([]*model.ChangelogEntry, error)
}
