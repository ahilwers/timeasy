package repository

import (
	"context"
	"github.com/gofrs/uuid"
	"timeasy-server/pkg/domain/model"
)

type UserRepository interface {
	GetUserProfile(ctx context.Context, userToken string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userToken string, updateRequest *model.UserProfileUpdateRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userToken string, passwordRequest *model.PasswordChangeRequest) error
	GetOrCreateUser(ctx context.Context, keycloakUserID uuid.UUID, keycloakData map[string]interface{}) (*model.User, error)
}