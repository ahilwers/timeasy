package repository

import (
	"timeasy-server/pkg/domain/model"
)

type SyncRepository interface {
	UpdateAndDeleteData(data model.SyncData) error
}
