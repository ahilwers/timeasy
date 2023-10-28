package repository

import (
	"time"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type SyncRepository interface {
	UpdateAndDeleteData(data model.SyncData) error
	GetUpdatedTimeEntriesOfUser(userId uuid.UUID, sinceWhen time.Time) ([]model.TimeEntry, error)
}
