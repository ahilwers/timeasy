package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
)

type GitLabProvider struct {
	config  *oauth2.Config
	client  *http.Client
	baseURL string
}

type GitLabIssue struct {
	ID        int    `json:"id"`
	IID       int    `json:"iid"`
	Title     string `json:"title"`
	State     string `json:"state"`
	WebURL    string `json:"web_url"`
	UpdatedAt string `json:"updated_at"`
}

type GitLabProject struct {
	ID                int    `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
}

func NewGitLabProvider(clientID, clientSecret, redirectURL, baseURL string) *GitLabProvider {
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"api"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  baseURL + "/oauth/authorize",
			TokenURL: baseURL + "/oauth/token",
		},
	}

	return &GitLabProvider{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

func (p *GitLabProvider) GetName() string {
	return "gitlab"
}

func (p *GitLabProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GitLabProvider) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *GitLabProvider) GetIssue(ctx context.Context, token string, projectRef string, keyOrNumber string) (*model.ExternalIssue, error) {
	// Parse issue number from keyOrNumber (e.g., "#123" -> "123")
	issueNumber := strings.TrimPrefix(keyOrNumber, "#")

	// URL encode the project reference
	encodedProject := url.PathEscape(projectRef)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/issues/%s", p.baseURL, encodedProject, issueNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
	}

	var glIssue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&glIssue); err != nil {
		return nil, err
	}

	return p.convertToExternalIssue(&glIssue, uuid.Nil), nil
}

func (p *GitLabProvider) ListIssues(ctx context.Context, token string, projectRef string) ([]*model.ExternalIssue, error) {
	encodedProject := url.PathEscape(projectRef)

	// Get open issues and recently updated closed issues
	states := []string{"opened", "closed"}
	var allIssues []*model.ExternalIssue

	since := time.Now().AddDate(0, 0, -90).Format(time.RFC3339)

	for _, state := range states {
		apiURL := fmt.Sprintf("%s/api/v4/projects/%s/issues?state=%s&per_page=100", p.baseURL, encodedProject, state)
		if state == "closed" {
			apiURL += "&updated_after=" + since
		}

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitLab API error: %d", resp.StatusCode)
		}

		var glIssues []GitLabIssue
		if err := json.NewDecoder(resp.Body).Decode(&glIssues); err != nil {
			return nil, err
		}

		for _, glIssue := range glIssues {
			allIssues = append(allIssues, p.convertToExternalIssue(&glIssue, uuid.Nil))
		}
	}

	return allIssues, nil
}

func (p *GitLabProvider) ValidateProjectRef(ctx context.Context, token string, projectRef string) error {
	encodedProject := url.PathEscape(projectRef)
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s", p.baseURL, encodedProject)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("project not accessible: %d", resp.StatusCode)
	}

	return nil
}

func (p *GitLabProvider) convertToExternalIssue(glIssue *GitLabIssue, projectID uuid.UUID) *model.ExternalIssue {
	updatedAt, _ := time.Parse(time.RFC3339, glIssue.UpdatedAt)

	return &model.ExternalIssue{
		ProjectID:   projectID,
		Provider:    p.GetName(),
		KeyOrNumber: fmt.Sprintf("#%d", glIssue.IID),
		Title:       glIssue.Title,
		State:       glIssue.State,
		URL:         glIssue.WebURL,
		UpdatedAt:   updatedAt,
	}
}

// IsGitLabIssuePattern checks if a string matches GitLab issue pattern (same as GitHub)
func IsGitLabIssuePattern(input string) bool {
	return IsGitHubIssuePattern(input)
}

// ExtractGitLabIssueNumber extracts the issue number from GitLab pattern (same as GitHub)
func ExtractGitLabIssueNumber(input string) string {
	return ExtractGitHubIssueNumber(input)
}
