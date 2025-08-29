package external

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"golang.org/x/oauth2"
)

// Helper function to construct Jira Basic Auth credentials
func constructJiraCredentials(token string, accountName string) string {
	if strings.Contains(token, ":") {
		// Token already contains email:api_token format (backward compatibility)
		return base64.StdEncoding.EncodeToString([]byte(token))
	} else if accountName != "" && strings.Contains(accountName, "@") {
		// Combine email (from accountName) and token
		authString := accountName + ":" + token
		return base64.StdEncoding.EncodeToString([]byte(authString))
	}
	// Fallback: assume token is already complete
	return base64.StdEncoding.EncodeToString([]byte(token))
}

type JiraProvider struct {
	config  *oauth2.Config
	client  *http.Client
	baseURL string
}

type JiraIssue struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Fields JiraFields `json:"fields"`
}

type JiraFields struct {
	Summary     string           `json:"summary"`
	Status      JiraStatus       `json:"status"`
	Updated     string           `json:"updated"`
}

type JiraStatus struct {
	Name string `json:"name"`
}

type JiraSearchResult struct {
	Issues []JiraIssue `json:"issues"`
}

type JiraProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

func NewJiraProvider(clientID, clientSecret, redirectURL, baseURL string) *JiraProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"read:jira-work", "read:jira-user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.atlassian.com/authorize",
			TokenURL: "https://auth.atlassian.com/oauth/token",
		},
	}

	return &JiraProvider{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}
}

func (p *JiraProvider) GetName() string {
	return "jira"
}

func (p *JiraProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, 
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("audience", "api.atlassian.com"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func (p *JiraProvider) ExchangeToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *JiraProvider) GetIssue(ctx context.Context, token string, projectRef string, keyOrNumber string) (*model.ExternalIssue, error) {
	apiURL := fmt.Sprintf("%s/rest/api/3/issue/%s", p.baseURL, keyOrNumber)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	
	// Use the same authentication method as in user and project validation
	if strings.Contains(token, ":") {
		// Token already contains email:api_token format
		credentials := base64.StdEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", "Basic "+credentials)
	} else {
		// Fallback: treat as Bearer token
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API error: %d", resp.StatusCode)
	}
	
	var jiraIssue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&jiraIssue); err != nil {
		return nil, err
	}
	
	return p.convertToExternalIssue(&jiraIssue, uuid.Nil), nil
}

func (p *JiraProvider) ListIssues(ctx context.Context, token string, projectRef string) ([]*model.ExternalIssue, error) {
	// Search for issues in the project
	jql := fmt.Sprintf("project = %s AND (status in (Open, \"In Progress\", \"To Do\") OR updated >= -90d) ORDER BY updated DESC", projectRef)
	
	apiURL := fmt.Sprintf("%s/rest/api/3/search", p.baseURL)
	
	reqBody := map[string]interface{}{
		"jql":        jql,
		"maxResults": 100,
		"fields": []string{"summary", "status", "updated"},
	}
	
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBytes)))
	if err != nil {
		return nil, err
	}
	
	// Use the same authentication method as in user and project validation
	if strings.Contains(token, ":") {
		// Token already contains email:api_token format
		credentials := base64.StdEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", "Basic "+credentials)
	} else {
		// Fallback: treat as Bearer token
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira API error: %d", resp.StatusCode)
	}
	
	var searchResult JiraSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}
	
	var allIssues []*model.ExternalIssue
	for _, jiraIssue := range searchResult.Issues {
		allIssues = append(allIssues, p.convertToExternalIssue(&jiraIssue, uuid.Nil))
	}
	
	return allIssues, nil
}

func (p *JiraProvider) ValidateProjectRef(ctx context.Context, token string, projectRef string) error {
	apiURL := fmt.Sprintf("%s/rest/api/3/project/%s", p.baseURL, projectRef)
	
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return err
	}
	
	// Use the same authentication method as in user validation
	// Jira uses Basic Authentication with email:token format
	if strings.Contains(token, ":") {
		// Token already contains email:api_token format
		credentials := base64.StdEncoding.EncodeToString([]byte(token))
		req.Header.Set("Authorization", "Basic "+credentials)
	} else {
		// Fallback: treat as Bearer token (though this likely won't work for Jira)
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	req.Header.Set("Accept", "application/json")
	
	slog.Debug("Jira project validation", "url", apiURL)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	slog.Debug("Jira project validation response", "status_code", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK {
		// Read response body for more detailed error
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		bodyStr := string(body[:n])
		slog.Debug("Jira project validation error response", "body", bodyStr, "status_code", resp.StatusCode)
		return fmt.Errorf("project not accessible: %d", resp.StatusCode)
	}
	
	return nil
}

func (p *JiraProvider) convertToExternalIssue(jiraIssue *JiraIssue, projectID uuid.UUID) *model.ExternalIssue {
	updatedAt, _ := time.Parse("2006-01-02T15:04:05.000-0700", jiraIssue.Fields.Updated)
	issueURL := fmt.Sprintf("%s/browse/%s", p.baseURL, jiraIssue.Key)
	
	return &model.ExternalIssue{
		ProjectID:   projectID,
		Provider:    "jira",
		KeyOrNumber: jiraIssue.Key,
		Title:       jiraIssue.Fields.Summary,
		State:       strings.ToLower(jiraIssue.Fields.Status.Name),
		URL:         issueURL,
		UpdatedAt:   updatedAt,
	}
}

// IsJiraIssuePattern checks if a string matches Jira issue pattern
func IsJiraIssuePattern(input string) bool {
	// Matches ABC-123 pattern
	pattern := `(?:^|\s)([A-Z][A-Z0-9]+-\d+)(?:\s|$)`
	re := regexp.MustCompile(pattern)
	return re.MatchString(input)
}

// ExtractJiraIssueKey extracts the issue key from Jira pattern
func ExtractJiraIssueKey(input string) string {
	pattern := `(?:^|\s)([A-Z][A-Z0-9]+-\d+)(?:\s|$)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}