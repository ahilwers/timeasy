package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type ExternalConnection struct {
	ID             uuid.UUID            `json:"id"`
	ProjectID      uuid.UUID            `json:"projectId"`
	UserAccountID  uuid.UUID            `json:"userAccountId"`
	UserAccount    *UserExternalAccount `json:"userAccount,omitempty"`
	Provider       string               `json:"provider"` // "github", "gitlab", "jira"
	ProjectRef     string               `json:"projectRef"` // "owner/repo" for GitHub/GitLab, "projectKey" for Jira
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

// ProviderType represents the external provider types
type ProviderType string

const (
	ProviderGitHub ProviderType = "github"
	ProviderGitLab ProviderType = "gitlab"
	ProviderJira   ProviderType = "jira"
)

// IsValidProvider checks if the provider is supported
func IsValidProvider(provider string) bool {
	switch ProviderType(provider) {
	case ProviderGitHub, ProviderGitLab, ProviderJira:
		return true
	default:
		return false
	}
}