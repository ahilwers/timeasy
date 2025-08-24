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
	// Validate that the user owns the project
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return fmt.Errorf("access denied: user does not own this project")
	}

	// Validate that the user owns the external account
	owned, err := uc.userAccountRepo.ValidateAccountOwnership(userAccountID, userID)
	if err != nil {
		return fmt.Errorf("failed to validate account ownership: %w", err)
	}
	if !owned {
		return fmt.Errorf("access denied: account not found or not owned by user")
	}

	// Get the user account to validate project reference
	account, err := uc.userAccountRepo.GetByID(userAccountID)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get provider instance and validate project reference
	providerInstance, exists := uc.providerFactory.GetProvider(account.Provider)
	if !exists {
		return fmt.Errorf("provider not configured: %s", account.Provider)
	}

	// For Jira, we need to create a temporary provider with the user's base URL and handle credentials
	if account.Provider == "jira" && account.BaseURL != "" {
		// Create a new Jira provider instance with the user's base URL
		jiraProvider := external.NewJiraProvider("dummy", "dummy", "", account.BaseURL)
		
		// Combine email and token for Jira authentication
		jiraToken := account.OAuthToken
		if !strings.Contains(jiraToken, ":") && account.AccountName != "" && strings.Contains(account.AccountName, "@") {
			jiraToken = account.AccountName + ":" + account.OAuthToken
		}
		
		if err := jiraProvider.ValidateProjectRef(ctx, jiraToken, projectRef); err != nil {
			return fmt.Errorf("invalid project reference: %w", err)
		}
	} else {
		if err := providerInstance.ValidateProjectRef(ctx, account.OAuthToken, projectRef); err != nil {
			return fmt.Errorf("invalid project reference: %w", err)
		}
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

	// Trigger initial sync
	go uc.SyncProjectIssues(context.Background(), projectID)

	return nil
}

// DisconnectProject removes an external connection for a project
func (uc *ExternalIntegrationUseCase) DisconnectProject(ctx context.Context, userID, projectID uuid.UUID) error {
	// Validate that the user owns the project
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return fmt.Errorf("access denied: user does not own this project")
	}

	// Get existing connection
	connection, err := uc.externalConnRepo.GetByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("no external connection found for project")
	}

	// Delete connection
	return uc.externalConnRepo.Delete(connection.ID)
}

// GetDescriptionSuggestions returns autocomplete suggestions for time entry descriptions
func (uc *ExternalIntegrationUseCase) GetDescriptionSuggestions(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]string, error) {
	// Validate that the user owns the project
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return nil, fmt.Errorf("access denied: user does not own this project")
	}

	// Get all time entries for the project
	timeEntries, err := uc.timeEntryRepo.GetAllTimeEntriesOfUserAndProject(userID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get time entries: %w", err)
	}

	// Extract unique descriptions that match the query
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
	fmt.Printf("DEBUG GetDescriptionSuggestions: Fetched %d external issues for project %v, error: %v\n", len(externalIssues), projectID, err)
	if err == nil { // Don't fail if no external issues exist
		for _, issue := range externalIssues {
			// Create suggestions in the format "#123 - Issue Title" (issue.KeyOrNumber already has #)
			issueRef := fmt.Sprintf("%s - %s", issue.KeyOrNumber, issue.Title)
			if strings.Contains(strings.ToLower(issueRef), queryLower) {
				descMap[issueRef] = 100 // Higher priority than regular descriptions
				fmt.Printf("DEBUG GetDescriptionSuggestions: Added issue suggestion: %s\n", issueRef)
			}
			
			// Only suggest just the issue number if the full description doesn't match
			// This prevents duplicates when both "#77" and "#77 - Title" would match
			simpleRef := issue.KeyOrNumber
			if strings.Contains(strings.ToLower(simpleRef), queryLower) && !strings.Contains(strings.ToLower(issueRef), queryLower) {
				descMap[simpleRef] = 99 // High priority
				fmt.Printf("DEBUG GetDescriptionSuggestions: Added simple issue suggestion: %s\n", simpleRef)
			}
		}
	} else {
		externalIssues = []*model.ExternalIssue{} // Empty slice for logging
	}

	// Sort by frequency (simple approach - convert to slice and return first N)
	suggestions := make([]string, 0) // Initialize empty slice instead of nil
	for desc := range descMap {
		suggestions = append(suggestions, desc)
		if len(suggestions) >= limit {
			break
		}
	}

	fmt.Printf("DEBUG GetDescriptionSuggestions: query='%s', found %d time entries, %d external issues, returning %d suggestions: %v\n", 
		query, len(timeEntries), len(externalIssues), len(suggestions), suggestions)

	return suggestions, nil
}

// ResolveIssue attempts to resolve an issue reference from the description
func (uc *ExternalIntegrationUseCase) ResolveIssue(ctx context.Context, userID, projectID uuid.UUID, input string) (*model.IssueResolveResult, error) {
	// Validate that the user owns the project
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or not accessible: %w", err)
	}

	if project.UserId != userID {
		return nil, fmt.Errorf("access denied: user does not own this project")
	}

	// Get external connection for the project (with user account)
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

	// Check if issue is already cached
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

	// Try to fetch from external provider
	provider, exists := uc.providerFactory.GetProvider(connection.Provider)
	if !exists {
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
	fmt.Printf("DEBUG: Starting sync for project %v\n", projectID)
	
	connection, err := uc.externalConnRepo.GetByProjectIDWithUserAccount(projectID)
	if err != nil {
		fmt.Printf("DEBUG: No external connection found for project %v: %v\n", projectID, err)
		return fmt.Errorf("no external connection found: %w", err)
	}
	fmt.Printf("DEBUG: Found connection for project %v: provider=%s, projectRef=%s\n", projectID, connection.Provider, connection.ProjectRef)
	if connection.UserAccount == nil {
		fmt.Printf("DEBUG: UserAccount is nil - no user external account data loaded!\n")
		return fmt.Errorf("no user account data found for external connection")
	}
	fmt.Printf("DEBUG: UserAccount found: ID=%s, AccountName=%s, HasToken=%v\n", 
		connection.UserAccount.ID, connection.UserAccount.AccountName, len(connection.UserAccount.OAuthToken) > 0)

	// For Jira, create a provider with the correct base URL
	var provider external.ExternalProvider
	var exists bool
	if connection.Provider == "jira" && connection.UserAccount.BaseURL != "" {
		fmt.Printf("DEBUG: Creating Jira provider with base URL: %s\n", connection.UserAccount.BaseURL)
		provider = external.NewJiraProvider("dummy", "dummy", "", connection.UserAccount.BaseURL)
		exists = true
	} else {
		provider, exists = uc.providerFactory.GetProvider(connection.Provider)
	}
	
	if !exists {
		fmt.Printf("DEBUG: Provider %s not configured\n", connection.Provider)
		return fmt.Errorf("provider not configured: %s", connection.Provider)
	}
	fmt.Printf("DEBUG: Found provider %s\n", connection.Provider)

	fmt.Printf("DEBUG: Fetching issues from %s for %s\n", connection.Provider, connection.ProjectRef)
	
	// For Jira, combine email and token if needed
	token := connection.UserAccount.OAuthToken
	if connection.Provider == "jira" && !strings.Contains(token, ":") && 
	   connection.UserAccount.AccountName != "" && strings.Contains(connection.UserAccount.AccountName, "@") {
		token = connection.UserAccount.AccountName + ":" + connection.UserAccount.OAuthToken
	}
	
	fmt.Printf("DEBUG: Using OAuth token: %s...\n", token[:10]) // Show first 10 chars only
	
	// Create a fresh context that won't be canceled
	freshCtx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(freshCtx, 30*time.Second)
	defer cancel()
	
	issues, err := provider.ListIssues(timeoutCtx, token, connection.ProjectRef)
	if err != nil {
		fmt.Printf("DEBUG: Failed to fetch issues: %v\n", err)
		return fmt.Errorf("failed to fetch issues: %w", err)
	}
	fmt.Printf("DEBUG: Fetched %d issues from %s\n", len(issues), connection.Provider)

	// Set project ID for all issues
	for _, issue := range issues {
		issue.ProjectID = projectID
	}

	// Batch upsert issues
	if err := uc.externalIssueRepo.BatchUpsert(issues); err != nil {
		fmt.Printf("DEBUG: Failed to upsert issues: %v\n", err)
		return fmt.Errorf("failed to upsert issues: %w", err)
	}
	fmt.Printf("DEBUG: Successfully upserted %d issues for project %v\n", len(issues), projectID)

	// Clean up old issues (older than 90 days)
	cutoffTime := time.Now().AddDate(0, 0, -90).Unix()
	if err := uc.externalIssueRepo.DeleteOldIssues(projectID, connection.Provider, cutoffTime); err != nil {
		fmt.Printf("Failed to clean up old issues: %v\n", err)
	}

	return nil
}

// ResolvePendingReferences resolves pending external references in time entries
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
