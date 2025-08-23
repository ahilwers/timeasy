package usecase

import (
	"context"
	"fmt"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
	"timeasy-server/pkg/external"

	"github.com/gofrs/uuid"
)

type UserExternalAccountUseCase struct {
	userAccountRepo repository.UserExternalAccountRepository
	providerFactory *external.ProviderFactory
}

func NewUserExternalAccountUseCase(
	userAccountRepo repository.UserExternalAccountRepository,
	providerFactory *external.ProviderFactory,
) *UserExternalAccountUseCase {
	return &UserExternalAccountUseCase{
		userAccountRepo: userAccountRepo,
		providerFactory: providerFactory,
	}
}

func (uc *UserExternalAccountUseCase) CreateAccount(ctx context.Context, userID uuid.UUID, req *model.UserExternalAccountRequest) (*model.UserExternalAccount, error) {
	if !model.IsValidProvider(req.Provider) {
		return nil, fmt.Errorf("unsupported provider: %s", req.Provider)
	}

	providerInstance, exists := uc.providerFactory.GetProvider(req.Provider)
	if !exists {
		return nil, fmt.Errorf("provider not configured: %s", req.Provider)
	}

	baseURL := req.BaseURL
	if baseURL == "" && req.Provider == "gitlab" {
		baseURL = "https://gitlab.com"
	}

	if err := uc.validateProviderConnection(ctx, providerInstance, req.OAuthToken, req.Provider, baseURL); err != nil {
		return nil, fmt.Errorf("invalid credentials or connection failed: %w", err)
	}

	account := &model.UserExternalAccount{
		UserID:      userID,
		Provider:    req.Provider,
		AccountName: req.AccountName,
		OAuthToken:  req.OAuthToken,
		BaseURL:     baseURL,
	}

	if err := uc.userAccountRepo.Create(account); err != nil {
		return nil, fmt.Errorf("failed to create external account: %w", err)
	}

	account.OAuthToken = ""
	return account, nil
}

func (uc *UserExternalAccountUseCase) UpdateAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, req *model.UserExternalAccountRequest) (*model.UserExternalAccount, error) {
	owned, err := uc.userAccountRepo.ValidateAccountOwnership(accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !owned {
		return nil, fmt.Errorf("access denied: account not found or not owned by user")
	}

	account, err := uc.userAccountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if account.Provider != req.Provider {
		return nil, fmt.Errorf("cannot change provider for existing account")
	}

	providerInstance, exists := uc.providerFactory.GetProvider(req.Provider)
	if !exists {
		return nil, fmt.Errorf("provider not configured: %s", req.Provider)
	}

	baseURL := req.BaseURL
	if baseURL == "" && req.Provider == "gitlab" {
		baseURL = "https://gitlab.com"
	}

	if err := uc.validateProviderConnection(ctx, providerInstance, req.OAuthToken, req.Provider, baseURL); err != nil {
		return nil, fmt.Errorf("invalid credentials or connection failed: %w", err)
	}

	account.AccountName = req.AccountName
	account.OAuthToken = req.OAuthToken
	account.BaseURL = baseURL

	if err := uc.userAccountRepo.Update(account); err != nil {
		return nil, fmt.Errorf("failed to update external account: %w", err)
	}

	// Remove sensitive data before returning
	account.OAuthToken = ""
	return account, nil
}

func (uc *UserExternalAccountUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error {
	owned, err := uc.userAccountRepo.ValidateAccountOwnership(accountID, userID)
	if err != nil {
		return fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !owned {
		return fmt.Errorf("access denied: account not found or not owned by user")
	}

	// TODO: Check if account is being used by any project connections
	// This would require checking the external_connections table

	if err := uc.userAccountRepo.Delete(accountID); err != nil {
		return fmt.Errorf("failed to delete external account: %w", err)
	}

	return nil
}

func (uc *UserExternalAccountUseCase) GetAccountsByUser(ctx context.Context, userID uuid.UUID) ([]*model.UserExternalAccount, error) {
	accounts, err := uc.userAccountRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	// Remove sensitive data
	for _, account := range accounts {
		account.OAuthToken = ""
	}

	return accounts, nil
}

func (uc *UserExternalAccountUseCase) GetAccountsByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) ([]*model.UserExternalAccount, error) {
	accounts, err := uc.userAccountRepo.GetByUserIDAndProvider(userID, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accounts: %w", err)
	}

	// Remove sensitive data
	for _, account := range accounts {
		account.OAuthToken = ""
	}

	return accounts, nil
}

func (uc *UserExternalAccountUseCase) TestAccount(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) error {
	owned, err := uc.userAccountRepo.ValidateAccountOwnership(accountID, userID)
	if err != nil {
		return fmt.Errorf("failed to validate ownership: %w", err)
	}
	if !owned {
		return fmt.Errorf("access denied: account not found or not owned by user")
	}

	account, err := uc.userAccountRepo.GetByID(accountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	providerInstance, exists := uc.providerFactory.GetProvider(account.Provider)
	if !exists {
		return fmt.Errorf("provider not configured: %s", account.Provider)
	}

	return uc.validateProviderConnection(ctx, providerInstance, account.OAuthToken, account.Provider, account.BaseURL)
}

func (uc *UserExternalAccountUseCase) validateProviderConnection(ctx context.Context, provider external.ExternalProvider, token string, providerType string, baseURL string) error {
	// For now, we'll use a simple validation by trying to fetch user info or a basic API call
	// This could be enhanced with provider-specific validation

	// For GitHub/GitLab, we could validate by trying to access user info
	// For Jira, we could validate by trying to access a basic endpoint

	switch providerType {
	case "github":
		// We could use the provider to validate, but for now just check if token is not empty
		if token == "" {
			return fmt.Errorf("empty token provided")
		}
	case "gitlab":
		// Similar validation
		if token == "" {
			return fmt.Errorf("empty token provided")
		}
	case "jira":
		// Similar validation
		if token == "" {
			return fmt.Errorf("empty token provided")
		}
		if baseURL == "" {
			return fmt.Errorf("base URL required for Jira")
		}
	default:
		return fmt.Errorf("unsupported provider: %s", providerType)
	}

	return nil
}
