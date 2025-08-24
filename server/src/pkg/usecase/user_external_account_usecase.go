package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

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

	// We don't need to get provider from factory anymore - validation is handled directly

	baseURL := req.BaseURL
	if baseURL == "" && req.Provider == "gitlab" {
		baseURL = "https://gitlab.com"
	}

	if err := uc.validateProviderConnectionWithAccountName(ctx, nil, req.OAuthToken, req.Provider, baseURL, req.AccountName); err != nil {
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

	// We don't need to get provider from factory anymore - validation is handled directly

	baseURL := req.BaseURL
	if baseURL == "" && req.Provider == "gitlab" {
		baseURL = "https://gitlab.com"
	}

	if err := uc.validateProviderConnectionWithAccountName(ctx, nil, req.OAuthToken, req.Provider, baseURL, req.AccountName); err != nil {
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

	// We don't need to get provider from factory anymore - validation is handled directly
	return uc.validateProviderConnectionWithAccountName(ctx, nil, account.OAuthToken, account.Provider, account.BaseURL, account.AccountName)
}

func (uc *UserExternalAccountUseCase) validateProviderConnection(ctx context.Context, provider external.ExternalProvider, token string, providerType string, baseURL string) error {
	return uc.validateProviderConnectionWithAccountName(ctx, provider, token, providerType, baseURL, "")
}

func (uc *UserExternalAccountUseCase) validateProviderConnectionWithAccountName(ctx context.Context, provider external.ExternalProvider, token string, providerType string, baseURL string, accountName string) error {
	if token == "" {
		return fmt.Errorf("empty token provided")
	}

	// For Jira, base URL is required
	if providerType == "jira" && baseURL == "" {
		return fmt.Errorf("base URL required for Jira")
	}

	// Test the actual connection by making a simple API call
	switch providerType {
	case "github":
		return uc.validateGitHubConnection(ctx, token)
	case "gitlab":
		return uc.validateGitLabConnection(ctx, token, baseURL)
	case "jira":
		return uc.validateJiraConnectionWithEmail(ctx, accountName, token, baseURL)
	default:
		return fmt.Errorf("unsupported provider: %s", providerType)
	}
}

func (uc *UserExternalAccountUseCase) validateGitHubConnection(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub API connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitHub authentication failed: invalid token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	return nil
}

func (uc *UserExternalAccountUseCase) validateGitLabConnection(ctx context.Context, token string, baseURL string) error {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	apiURL := baseURL + "/api/v4/user"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GitLab API connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitLab authentication failed: invalid token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API error: status %d", resp.StatusCode)
	}

	return nil
}

func (uc *UserExternalAccountUseCase) validateJiraConnection(ctx context.Context, token string, baseURL string) error {
	// This method needs access to the account name (email) to construct proper credentials
	// For now, we'll assume the token is just the API token and we'll handle the email:token combination in the provider methods
	return uc.validateJiraConnectionWithEmail(ctx, "", token, baseURL)
}

func (uc *UserExternalAccountUseCase) validateJiraConnectionWithEmail(ctx context.Context, email string, token string, baseURL string) error {
	// Try the older API version first as it might be more compatible
	apiURL := baseURL + "/rest/api/2/myself"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set up Basic Auth credentials
	if email != "" && !strings.Contains(token, ":") {
		// Combine email and token
		authString := email + ":" + token
		credentials := base64.StdEncoding.EncodeToString([]byte(authString))
		req.Header.Set("Authorization", "Basic "+credentials)
		fmt.Printf("DEBUG: Using Basic Auth with email:token combination\n")
	} else if strings.Contains(token, ":") {
		// Token already contains email:api_token format (backward compatibility)
		credentials := base64.StdEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", "Basic "+credentials)
		fmt.Printf("DEBUG: Using Basic Auth with existing email:token format\n")
	} else {
		// Fallback: try Bearer token
		req.Header.Set("Authorization", "Bearer "+token)
		fmt.Printf("DEBUG: Using Bearer token as fallback\n")
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Jira API connection failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: Jira validation - URL: %s, Status: %d\n", apiURL, resp.StatusCode)
	fmt.Printf("DEBUG: Jira validation - Email: %s, Token length: %d\n", email, len(token))
	if email != "" && !strings.Contains(token, ":") {
		fmt.Printf("DEBUG: Jira validation - Using combined credentials: %s:[token]\n", email)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Jira authentication failed: Invalid email or API token")
	}
	if resp.StatusCode != http.StatusOK {
		// Read response body for more detailed error
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])
		fmt.Printf("DEBUG: Jira error response body: %s\n", bodyStr)
		return fmt.Errorf("Jira API error: status %d - %s", resp.StatusCode, bodyStr)
	}

	return nil
}
