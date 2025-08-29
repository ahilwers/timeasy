package external

import (
	"context"
	"fmt"
	"timeasy-server/pkg/domain/model"

	"golang.org/x/oauth2"
)

// ExternalProvider defines the interface for external project providers
type ExternalProvider interface {
	// GetName returns the provider name (github, gitlab, jira)
	GetName() string
	
	// GetAuthURL returns the OAuth2 authorization URL
	GetAuthURL(state string) string
	
	// ExchangeToken exchanges the authorization code for an access token
	ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error)
	
	// GetIssue fetches a specific issue by its key/number
	GetIssue(ctx context.Context, token string, projectRef string, keyOrNumber string) (*model.ExternalIssue, error)
	
	// ListIssues fetches issues from the external provider
	ListIssues(ctx context.Context, token string, projectRef string) ([]*model.ExternalIssue, error)
	
	// ValidateProjectRef validates if the project reference is accessible
	ValidateProjectRef(ctx context.Context, token string, projectRef string) error
}

// ProviderFactory creates provider instances
type ProviderFactory struct {
	providers map[string]ExternalProvider
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]ExternalProvider),
	}
}

// RegisterProvider registers a provider
func (f *ProviderFactory) RegisterProvider(provider ExternalProvider) {
	f.providers[provider.GetName()] = provider
}

// GetProvider returns a provider by name
func (f *ProviderFactory) GetProvider(name string) (ExternalProvider, bool) {
	provider, exists := f.providers[name]
	return provider, exists
}

// GetAllProviders returns all registered providers
func (f *ProviderFactory) GetAllProviders() map[string]ExternalProvider {
	return f.providers
}

// CreateProvider creates a provider instance based on provider name and configuration
func (f *ProviderFactory) CreateProvider(providerName string, baseURL string) (ExternalProvider, error) {
	switch providerName {
	case "github":
		return NewGitHubProvider("", "", ""), nil
	case "gitlab":
		if baseURL == "" {
			baseURL = "https://gitlab.com"
		}
		return NewGitLabProvider("", "", "", baseURL), nil
	case "jira":
		if baseURL == "" {
			return nil, fmt.Errorf("base URL is required for Jira provider")
		}
		return NewJiraProvider("", "", "", baseURL), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}