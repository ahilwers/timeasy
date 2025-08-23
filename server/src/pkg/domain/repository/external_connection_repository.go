package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type ExternalConnectionRepository interface {
	Create(connection *model.ExternalConnection) error
	Update(connection *model.ExternalConnection) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*model.ExternalConnection, error)
	GetByProjectID(projectID uuid.UUID) (*model.ExternalConnection, error)
	GetByProjectIDAndProvider(projectID uuid.UUID, provider string) (*model.ExternalConnection, error)
	List(userID uuid.UUID) ([]*model.ExternalConnection, error)
	GetWithUserAccount(id uuid.UUID) (*model.ExternalConnection, error)
	GetByProjectIDWithUserAccount(projectID uuid.UUID) (*model.ExternalConnection, error)
}