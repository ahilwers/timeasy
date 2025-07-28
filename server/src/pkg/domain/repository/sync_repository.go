package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type SyncRepository interface {
	UpdateAndDeleteData(data model.SyncData) error
	GetUpdatedTimeEntriesOfUser(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.TimeEntrySyncResult, error)

	GetUpdatedProjectsOfUser(userId uuid.UUID, sinceTimeLogEntry int64, excludeClientId string) (model.ProjectSyncResult, error)
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetTimeEntryById(id uuid.UUID) (*model.TimeEntry, error)
}
