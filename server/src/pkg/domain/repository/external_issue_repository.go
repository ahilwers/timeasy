package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type ExternalIssueRepository interface {
	Create(issue *model.ExternalIssue) error
	Update(issue *model.ExternalIssue) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*model.ExternalIssue, error)
	GetByProjectIDAndKey(projectID uuid.UUID, provider string, keyOrNumber string) (*model.ExternalIssue, error)
	List(projectID uuid.UUID) ([]*model.ExternalIssue, error)
	ListByProvider(projectID uuid.UUID, provider string) ([]*model.ExternalIssue, error)
	BatchUpsert(issues []*model.ExternalIssue) error
	DeleteOldIssues(projectID uuid.UUID, provider string, cutoffTime int64) error
}