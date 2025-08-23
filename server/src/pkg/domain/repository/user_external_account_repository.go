package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type UserExternalAccountRepository interface {
	Create(account *model.UserExternalAccount) error
	Update(account *model.UserExternalAccount) error
	Delete(id uuid.UUID) error
	GetByID(id uuid.UUID) (*model.UserExternalAccount, error)
	GetByUserIDAndProvider(userID uuid.UUID, provider string) ([]*model.UserExternalAccount, error)
	GetByUserID(userID uuid.UUID) ([]*model.UserExternalAccount, error)
	ValidateAccountOwnership(accountID uuid.UUID, userID uuid.UUID) (bool, error)
}