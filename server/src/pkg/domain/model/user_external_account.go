package model

import (
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type UserExternalAccount struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	Provider    string    `json:"provider"` // "github", "gitlab", "jira"
	AccountName string    `json:"accountName"` // Display name for the account
	OAuthToken  string    `json:"-"` // Never expose in JSON
	BaseURL     string    `json:"baseURL,omitempty"` // For GitLab self-hosted or Jira instances
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// GetAuthToken returns the properly formatted authentication token for API calls.
// For Jira, it combines the email (AccountName) with the API token if needed.
// For other providers (GitHub, GitLab), it returns the token as-is.
func (u *UserExternalAccount) GetAuthToken() string {
	if u.Provider == "jira" && !strings.Contains(u.OAuthToken, ":") && 
		u.AccountName != "" && strings.Contains(u.AccountName, "@") {
		return u.AccountName + ":" + u.OAuthToken
	}
	return u.OAuthToken
}

// UserExternalAccountRequest represents the request to create/update an external account
type UserExternalAccountRequest struct {
	Provider    string `json:"provider" binding:"required"`
	AccountName string `json:"accountName" binding:"required"`
	OAuthToken  string `json:"oauthToken" binding:"required"`
	BaseURL     string `json:"baseURL,omitempty"`
}

// ConnectProjectToAccountRequest represents the request to link a project to an external account
type ConnectProjectToAccountRequest struct {
	UserAccountID uuid.UUID `json:"userAccountId" binding:"required"`
	ProjectRef    string    `json:"projectRef" binding:"required"`
}