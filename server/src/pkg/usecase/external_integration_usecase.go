package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
	"timeasy-server/pkg/external"

	"github.com/gofrs/uuid"
)

type ExternalIntegrationUseCase struct {
	externalConnRepo  repository.ExternalConnectionRepository
	externalIssueRepo repository.ExternalIssueRepository
	userAccountRepo   repository.UserExternalAccountRepository
	timeEntryRepo     repository.TimeEntryRepository
	projectRepo       repository.ProjectRepository
	providerFactory   *external.ProviderFactory
}

func NewExternalIntegrationUseCase(
	externalConnRepo repository.ExternalConnectionRepository,
	externalIssueRepo repository.ExternalIssueRepository,
	userAccountRepo repository.UserExternalAccountRepository,
	timeEntryRepo repository.TimeEntryRepository,
	projectRepo repository.ProjectRepository,
	providerFactory *external.ProviderFactory,
) *ExternalIntegrationUseCase {
	return &ExternalIntegrationUseCase{
		externalConnRepo:  externalConnRepo,
		externalIssueRepo: externalIssueRepo,
		userAccountRepo:   userAccountRepo,
		timeEntryRepo:     timeEntryRepo,
		projectRepo:       projectRepo,
		providerFactory:   providerFactory,
	}
}

// ConnectProjectToAccount creates an external connection between a project and a user's external account
func (uc *ExternalIntegrationUseCase) ConnectProjectToAccount(ctx context.Context, userID, projectID, userAccountID uuid.UUID, projectRef string) error {
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return fmt.Errorf("access denied: user does not own this project")
	}

	owned, err := uc.userAccountRepo.ValidateAccountOwnership(userAccountID, userID)
	if err != nil {
		return fmt.Errorf("failed to validate account ownership: %w", err)
	}
	if !owned {
		return fmt.Errorf("access denied: account not found or not owned by user")
	}

	account, err := uc.userAccountRepo.GetByID(userAccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	var providerInstance external.ExternalProvider
	switch account.Provider {
	case "github":
		providerInstance = external.NewGitHubProvider("", "", "")
	case "gitlab":
		baseURL := account.BaseURL
		if baseURL == "" {
			baseURL = "https://gitlab.com"
		}
		providerInstance = external.NewGitLabProvider("", "", "", baseURL)
	case "jira":
		if account.BaseURL == "" {
			return fmt.Errorf("base URL required for Jira")
		}
		providerInstance = external.NewJiraProvider("", "", "", account.BaseURL)
	default:
		return fmt.Errorf("unsupported provider: %s", account.Provider)
	}

	token := account.OAuthToken
	if account.Provider == "jira" && !strings.Contains(token, ":") && account.AccountName != "" && strings.Contains(account.AccountName, "@") {
		token = account.AccountName + ":" + account.OAuthToken
	}

	if err := providerInstance.ValidateProjectRef(ctx, token, projectRef); err != nil {
		return fmt.Errorf("invalid project reference: %w", err)
	}

	// Check if connection already exists for this project
	existingConn, err := uc.externalConnRepo.GetByProjectID(projectID)
	if err == nil && existingConn != nil {
		// Update existing connection
		existingConn.UserAccountID = userAccountID
		existingConn.Provider = account.Provider
		existingConn.ProjectRef = projectRef
		if err := uc.externalConnRepo.Update(existingConn); err != nil {
			return fmt.Errorf("failed to update connection: %w", err)
		}
	} else {
		// Create new connection
		connection := &model.ExternalConnection{
			ProjectID:     projectID,
			UserAccountID: userAccountID,
			Provider:      account.Provider,
			ProjectRef:    projectRef,
		}

		if err := uc.externalConnRepo.Create(connection); err != nil {
			return fmt.Errorf("failed to create external connection: %w", err)
		}
	}

	go uc.SyncProjectIssues(context.Background(), projectID)
	return nil
}

// DisconnectProject removes an external connection for a project
func (uc *ExternalIntegrationUseCase) DisconnectProject(ctx context.Context, userID, projectID uuid.UUID) error {
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return fmt.Errorf("access denied: user does not own this project")
	}

	connection, err := uc.externalConnRepo.GetByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("no external connection found for project")
	}

	return uc.externalConnRepo.Delete(connection.ID)
}

// GetDescriptionSuggestions returns autocomplete suggestions for time entry descriptions
func (uc *ExternalIntegrationUseCase) GetDescriptionSuggestions(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]string, error) {
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return nil, fmt.Errorf("access denied: user does not own this project")
	}

	timeEntries, err := uc.timeEntryRepo.GetAllTimeEntriesOfUserAndProject(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries: %w", err)
	}

	descMap := make(map[string]int)
	queryLower := strings.ToLower(query)

	// Add suggestions from existing time entry descriptions
	for _, entry := range timeEntries {
		desc := strings.TrimSpace(entry.Description)
		if desc != "" && strings.Contains(strings.ToLower(desc), queryLower) {
			descMap[desc]++
		}
	}

	// Also add suggestions from external issues (GitHub, GitLab, Jira issues)
	var externalIssues []*model.ExternalIssue
	externalIssues, err = uc.externalIssueRepo.List(projectID)
	if err == nil { // Don't fail if no external issues exist
		for _, issue := range externalIssues {
			// Create suggestions in the format "#123 - Issue Title" (issue.KeyOrNumber already has #)
			issueRef := fmt.Sprintf("%s - %s", issue.KeyOrNumber, issue.Title)
			if strings.Contains(strings.ToLower(issueRef), queryLower) {
				descMap[issueRef] = 100 // Higher priority than regular descriptions
			}

			// Only suggest just the issue number if the full description doesn't match
			// This prevents duplicates when both "#77" and "#77 - Title" would match
			simpleRef := issue.KeyOrNumber
			if strings.Contains(strings.ToLower(simpleRef), queryLower) && !strings.Contains(strings.ToLower(issueRef), queryLower) {
				descMap[simpleRef] = 99 // High priority
			}
		}
	} else {
		externalIssues = []*model.ExternalIssue{} // Empty slice for logging
	}

	suggestions := make([]string, 0)
	for desc := range descMap {
		suggestions = append(suggestions, desc)
		if len(suggestions) >= limit {
			break
		}
	}
	return suggestions, nil
}

// ResolveIssue attempts to resolve an issue reference from the description
func (uc *ExternalIntegrationUseCase) ResolveIssue(ctx context.Context, userID, projectID uuid.UUID, input string) (*model.IssueResolveResult, error) {
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return nil, fmt.Errorf("access denied: user does not own this project")
	}

	connection, err := uc.externalConnRepo.GetByProjectIDWithUserAccount(projectID)
	if err != nil {
		// No external connection, return pending
		return &model.IssueResolveResult{
			Key:    input,
			Status: "pending",
		}, nil
	}

	// Extract issue key/number based on provider
	var keyOrNumber string
	switch connection.Provider {
	case "github", "gitlab":
		if external.IsGitHubIssuePattern(input) {
			keyOrNumber = external.ExtractGitHubIssueNumber(input)
		}
	case "jira":
		if external.IsJiraIssuePattern(input) {
			keyOrNumber = external.ExtractJiraIssueKey(input)
		}
	}

	if keyOrNumber == "" {
		return &model.IssueResolveResult{
			Key:    input,
			Status: "pending",
		}, nil
	}

	cachedIssue, err := uc.externalIssueRepo.GetByProjectIDAndKey(projectID, connection.Provider, keyOrNumber)
	if err == nil && cachedIssue != nil {
		return &model.IssueResolveResult{
			IssueID: &cachedIssue.ID,
			Key:     cachedIssue.KeyOrNumber,
			Title:   cachedIssue.Title,
			State:   cachedIssue.State,
			URL:     cachedIssue.URL,
			Status:  "resolved",
		}, nil
	}

	var provider external.ExternalProvider
	switch connection.Provider {
	case "github":
		provider = external.NewGitHubProvider("", "", "")
	case "gitlab":
		baseURL := connection.UserAccount.BaseURL
		if baseURL == "" {
			baseURL = "https://gitlab.com"
		}
		provider = external.NewGitLabProvider("", "", "", baseURL)
	case "jira":
		if connection.UserAccount.BaseURL == "" {
			return &model.IssueResolveResult{
				Key:    keyOrNumber,
				Status: "pending",
			}, nil
		}
		provider = external.NewJiraProvider("", "", "", connection.UserAccount.BaseURL)
	default:
		return &model.IssueResolveResult{
			Key:    keyOrNumber,
			Status: "pending",
		}, nil
	}

	// For Jira, combine email and token if needed
	token := connection.UserAccount.OAuthToken
	if connection.Provider == "jira" && !strings.Contains(token, ":") &&
		connection.UserAccount.AccountName != "" && strings.Contains(connection.UserAccount.AccountName, "@") {
		token = connection.UserAccount.AccountName + ":" + connection.UserAccount.OAuthToken
	}

	externalIssue, err := provider.GetIssue(ctx, token, connection.ProjectRef, keyOrNumber)
	if err != nil {
		return &model.IssueResolveResult{
			Key:    keyOrNumber,
			Status: "pending",
		}, nil
	}

	// Cache the issue
	externalIssue.ProjectID = projectID
	if err := uc.externalIssueRepo.Create(externalIssue); err != nil {
		// Log error but continue
		fmt.Printf("Failed to cache external issue: %v\n", err)
	}

	return &model.IssueResolveResult{
		IssueID: &externalIssue.ID,
		Key:     externalIssue.KeyOrNumber,
		Title:   externalIssue.Title,
		State:   externalIssue.State,
		URL:     externalIssue.URL,
		Status:  "resolved",
	}, nil
}

// SyncProjectIssues synchronizes issues from external provider
func (uc *ExternalIntegrationUseCase) SyncProjectIssues(ctx context.Context, projectID uuid.UUID) error {
	connection, err := uc.externalConnRepo.GetByProjectIDWithUserAccount(projectID)
	if err != nil {
		return fmt.Errorf("no external connection found: %w", err)
	}
	if connection.UserAccount == nil {
		return fmt.Errorf("no user account data found for external connection")
	}

	// Create provider instance dynamically based on connection settings
	var provider external.ExternalProvider
	switch connection.Provider {
	case "github":
		provider = external.NewGitHubProvider("", "", "")
	case "gitlab":
		baseURL := connection.UserAccount.BaseURL
		if baseURL == "" {
			baseURL = "https://gitlab.com"
		}
		provider = external.NewGitLabProvider("", "", "", baseURL)
	case "jira":
		if connection.UserAccount.BaseURL == "" {
			return fmt.Errorf("base URL required for Jira")
		}
		provider = external.NewJiraProvider("", "", "", connection.UserAccount.BaseURL)
	default:
		return fmt.Errorf("unsupported provider: %s", connection.Provider)
	}

	// For Jira, combine email and token if needed
	token := connection.UserAccount.OAuthToken
	if connection.Provider == "jira" && !strings.Contains(token, ":") &&
		connection.UserAccount.AccountName != "" && strings.Contains(connection.UserAccount.AccountName, "@") {
		token = connection.UserAccount.AccountName + ":" + connection.UserAccount.OAuthToken
	}

	// Create a fresh context that won't be canceled
	freshCtx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(freshCtx, 30*time.Second)
	defer cancel()

	issues, err := provider.ListIssues(timeoutCtx, token, connection.ProjectRef)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	// Set project ID for all issues
	for _, issue := range issues {
		issue.ProjectID = projectID
	}

	if err := uc.externalIssueRepo.BatchUpsert(issues); err != nil {
		return fmt.Errorf("failed to upsert issues: %w", err)
	}

	// Clean up old issues (older than 90 days)
	cutoffTime := time.Now().AddDate(0, 0, -90).Unix()
	if err := uc.externalIssueRepo.DeleteOldIssues(projectID, connection.Provider, cutoffTime); err != nil {
		fmt.Printf("Failed to clean up old issues: %v\n", err)
	}

	return nil
}

func (uc *ExternalIntegrationUseCase) ResolvePendingReferences(ctx context.Context, userID uuid.UUID) error {
	// Get all time entries with pending external references
	allEntries, err := uc.timeEntryRepo.GetAllTimeEntriesOfUser(userID)
	if err != nil {
		return fmt.Errorf("failed to get time entries: %w", err)
	}

	tx, err := uc.timeEntryRepo.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, entry := range allEntries {
		if entry.PendingExternalRef == nil || *entry.PendingExternalRef == "" {
			continue
		}

		// Try to resolve the pending reference
		result, err := uc.ResolveIssue(ctx, userID, entry.ProjectId, *entry.PendingExternalRef)
		if err != nil {
			continue // Skip this entry, log error if needed
		}

		if result.Status == "resolved" && result.IssueID != nil {
			// Update the time entry
			entry.ExternalIssueID = result.IssueID
			entry.PendingExternalRef = nil // Clear pending reference

			if err := uc.timeEntryRepo.UpdateTimeEntry(&entry, tx); err != nil {
				fmt.Printf("Failed to update time entry %s: %v\n", entry.ID, err)
			}
		}
	}

	return tx.Commit()
}

// DetectIssuePattern detects if input contains issue patterns
func (uc *ExternalIntegrationUseCase) DetectIssuePattern(input string) (string, string) {
	// GitHub/GitLab pattern: #123 or owner/repo#123
	if external.IsGitHubIssuePattern(input) {
		return external.ExtractGitHubIssueNumber(input), "github"
	}

	// Jira pattern: ABC-123
	if external.IsJiraIssuePattern(input) {
		return external.ExtractJiraIssueKey(input), "jira"
	}

	return "", ""
}

func (uc *ExternalIntegrationUseCase) GetExternalIssueByID(ctx context.Context, issueID uuid.UUID) (*model.ExternalIssue, error) {
	return uc.externalIssueRepo.GetByID(issueID)
}
