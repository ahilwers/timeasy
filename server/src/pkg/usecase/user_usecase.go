package usecase

import (
	"context"
	"fmt"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
)

type UserUsecase interface {
	GetUserProfile(ctx context.Context, userToken string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userToken string, updateRequest *model.UserProfileUpdateRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userToken string, passwordRequest *model.PasswordChangeRequest) error
}

type userUsecase struct {
	userRepository repository.UserRepository
}

func NewUserUsecase(userRepository repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepository: userRepository,
	}
}

func (u *userUsecase) GetUserProfile(ctx context.Context, userToken string) (*model.User, error) {
	user, err := u.userRepository.GetUserProfile(ctx, userToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return user, nil
}

func (u *userUsecase) UpdateUserProfile(ctx context.Context, userToken string, updateRequest *model.UserProfileUpdateRequest) (*model.User, error) {
	if updateRequest.Email != nil && *updateRequest.Email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	user, err := u.userRepository.UpdateUserProfile(ctx, userToken, updateRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}
	return user, nil
}

func (u *userUsecase) ChangePassword(ctx context.Context, userToken string, passwordRequest *model.PasswordChangeRequest) error {
	if len(passwordRequest.NewPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters long")
	}

	err := u.userRepository.ChangePassword(ctx, userToken, passwordRequest)
	if err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}
	return nil
}
