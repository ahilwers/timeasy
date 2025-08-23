package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
)

type GitHubProvider struct {
	config *oauth2.Config
	client *http.Client
}

type GitHubIssue struct {
	ID       int    `json:"id"`
	Number   int    `json:"number"`
	Title    string `json:"title"`
	State    string `json:"state"`
	HTMLURL  string `json:"html_url"`
	UpdatedAt string `json:"updated_at"`
}

type GitHubRepo struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

func NewGitHubProvider(clientID, clientSecret, redirectURL string) *GitHubProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	return &GitHubProvider{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *GitHubProvider) GetName() string {
	return "github"
}

func (p *GitHubProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GitHubProvider) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *GitHubProvider) GetIssue(ctx context.Context, token string, projectRef string, keyOrNumber string) (*model.ExternalIssue, error) {
	// Parse issue number from keyOrNumber (e.g., "#123" -> "123")
	issueNumber := strings.TrimPrefix(keyOrNumber, "#")
	
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", projectRef, issueNumber)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}
	
	var ghIssue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&ghIssue); err != nil {
		return nil, err
	}
	
	return p.convertToExternalIssue(&ghIssue, uuid.Nil), nil
}

func (p *GitHubProvider) ListIssues(ctx context.Context, token string, projectRef string) ([]*model.ExternalIssue, error) {
	// Get open issues and recently updated closed issues
	urls := []string{
		fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=100", projectRef),
		fmt.Sprintf("https://api.github.com/repos/%s/issues?state=closed&since=%s&per_page=100", 
			projectRef, time.Now().AddDate(0, 0, -90).Format(time.RFC3339)),
	}
	
	var allIssues []*model.ExternalIssue
	
	for _, url := range urls {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		
		resp, err := p.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
		}
		
		var ghIssues []GitHubIssue
		if err := json.NewDecoder(resp.Body).Decode(&ghIssues); err != nil {
			return nil, err
		}
		
		for _, ghIssue := range ghIssues {
			allIssues = append(allIssues, p.convertToExternalIssue(&ghIssue, uuid.Nil))
		}
	}
	
	return allIssues, nil
}

func (p *GitHubProvider) ValidateProjectRef(ctx context.Context, token string, projectRef string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s", projectRef)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("repository not accessible: %d", resp.StatusCode)
	}
	
	return nil
}

func (p *GitHubProvider) convertToExternalIssue(ghIssue *GitHubIssue, projectID uuid.UUID) *model.ExternalIssue {
	updatedAt, _ := time.Parse(time.RFC3339, ghIssue.UpdatedAt)
	
	return &model.ExternalIssue{
		ProjectID:   projectID,
		Provider:    p.GetName(),
		KeyOrNumber: fmt.Sprintf("#%d", ghIssue.Number),
		Title:       ghIssue.Title,
		State:       ghIssue.State,
		URL:         ghIssue.HTMLURL,
		UpdatedAt:   updatedAt,
	}
}

// IsGitHubIssuePattern checks if a string matches GitHub issue pattern
func IsGitHubIssuePattern(input string) bool {
	// Matches #123 or owner/repo#123
	pattern := `(?:^|\s)(?:[\w.-]+/[\w.-]+)?#(\d+)(?:\s|$)`
	re := regexp.MustCompile(pattern)
	return re.MatchString(input)
}

// ExtractGitHubIssueNumber extracts the issue number from GitHub pattern
func ExtractGitHubIssueNumber(input string) string {
	pattern := `(?:^|\s)(?:[\w.-]+/[\w.-]+)?#(\d+)(?:\s|$)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return "#" + matches[1]
	}
	return ""
}