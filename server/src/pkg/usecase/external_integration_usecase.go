package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
	"timeasy-server/pkg/external"

	"github.com/gofrs/uuid"
)

// SyncScheduler interface for managing project sync scheduling
type SyncScheduler interface {
	RegisterProject(projectID uuid.UUID)
	UnregisterProject(projectID uuid.UUID)
	TriggerImmediateSync(projectID uuid.UUID)
}

type ExternalIntegrationUseCase struct {
	externalConnRepo  repository.ExternalConnectionRepository
	externalIssueRepo repository.ExternalIssueRepository
	userAccountRepo   repository.UserExternalAccountRepository
	timeEntryRepo     repository.TimeEntryRepository
	projectRepo       repository.ProjectRepository
	providerFactory   *external.ProviderFactory
	rateLimiter       *external.ProviderRateLimiter
	teamUsecase       TeamUsecase
	syncScheduler     SyncScheduler
}

func NewExternalIntegrationUseCase(
	externalConnRepo repository.ExternalConnectionRepository,
	externalIssueRepo repository.ExternalIssueRepository,
	userAccountRepo repository.UserExternalAccountRepository,
	timeEntryRepo repository.TimeEntryRepository,
	projectRepo repository.ProjectRepository,
	providerFactory *external.ProviderFactory,
	teamUsecase TeamUsecase,
) *ExternalIntegrationUseCase {
	return &ExternalIntegrationUseCase{
		externalConnRepo:  externalConnRepo,
		externalIssueRepo: externalIssueRepo,
		userAccountRepo:   userAccountRepo,
		timeEntryRepo:     timeEntryRepo,
		projectRepo:       projectRepo,
		providerFactory:   providerFactory,
		rateLimiter:       external.NewProviderRateLimiter(),
		teamUsecase:       teamUsecase,
		syncScheduler:     nil,
	}
}

func (uc *ExternalIntegrationUseCase) SetSyncScheduler(scheduler SyncScheduler) {
	uc.syncScheduler = scheduler
}

// ConnectProjectToAccount creates an external connection between a project and a user's external account
func (uc *ExternalIntegrationUseCase) ConnectProjectToAccount(ctx context.Context, userID, projectID, userAccountID uuid.UUID, projectRef string) error {
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return fmt.Errorf("project not found or not accessible: %w", err)
	}

	if !uc.isUserProjectAdmin(userID, project) {
		return fmt.Errorf("access denied: user does not have admin access to this project")
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

	providerInstance, err := uc.providerFactory.CreateProvider(account.Provider, account.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	token := account.GetAuthToken()

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

	if uc.syncScheduler != nil {
		uc.syncScheduler.RegisterProject(projectID)
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

	if !uc.isUserProjectAdmin(userID, project) {
		return fmt.Errorf("access denied: user does not have admin access to this project")
	}

	connection, err := uc.externalConnRepo.GetByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("no external connection found for project")
	}

	if uc.syncScheduler != nil {
		uc.syncScheduler.UnregisterProject(projectID)
	}

	return uc.externalConnRepo.Delete(connection.ID)
}

// GetDescriptionSuggestions returns autocomplete suggestions for time entry descriptions
func (uc *ExternalIntegrationUseCase) GetDescriptionSuggestions(ctx context.Context, userID, projectID uuid.UUID, query string, limit int) ([]string, error) {
	project, err := uc.projectRepo.GetProjectById(projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found or not accessible: %w", err)
	}

	if !uc.doesProjectBelongToUser(userID, project) {
		return nil, fmt.Errorf("access denied: user does not have access to this project")
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

	if !uc.doesProjectBelongToUser(userID, project) {
		return nil, fmt.Errorf("access denied: user does not have access to this project")
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

	provider, err := uc.providerFactory.CreateProvider(connection.Provider, connection.UserAccount.BaseURL)
	if err != nil {
		return &model.IssueResolveResult{
			Key:    keyOrNumber,
			Status: "pending",
		}, nil
	}

	token := connection.UserAccount.GetAuthToken()

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
		slog.Warn("Failed to cache external issue", "error", err, "issue_id", externalIssue.ID)
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

	provider, err := uc.providerFactory.CreateProvider(connection.Provider, connection.UserAccount.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	token := connection.UserAccount.GetAuthToken()

	// Create a fresh context that won't be canceled
	freshCtx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(freshCtx, 30*time.Second)
	defer cancel()

	// Apply rate limiting before making API calls
	if err := uc.rateLimiter.WaitForPermission(timeoutCtx, connection.Provider); err != nil {
		return fmt.Errorf("rate limiting error: %w", err)
	}

	issues, err := provider.ListIssues(timeoutCtx, token, connection.ProjectRef)
	if err != nil {
		// Record the failure with appropriate type
		isRateLimit := strings.Contains(err.Error(), "rate limit") ||
			strings.Contains(err.Error(), "429") ||
			strings.Contains(err.Error(), "too many requests")
		uc.rateLimiter.RecordFailure(connection.Provider, isRateLimit)
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	// Record successful API call
	uc.rateLimiter.RecordSuccess(connection.Provider)

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
		slog.Warn("Failed to clean up old issues", "error", err, "project_id", projectID, "provider", connection.Provider)
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
				slog.Error("Failed to update time entry", "entry_id", entry.ID, "error", err)
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

// ProcessTimeEntryIssueDetection detects and resolves issues in time entry descriptions
func (uc *ExternalIntegrationUseCase) ProcessTimeEntryIssueDetection(ctx context.Context, timeEntry *model.TimeEntry, userId uuid.UUID) error {
	// Skip if description is empty
	if timeEntry.Description == "" {
		return nil
	}

	// Detect issue patterns
	key, provider := uc.DetectIssuePattern(timeEntry.Description)
	if key == "" || provider == "" {
		// No issue pattern detected, clear any existing external issue fields
		timeEntry.ExternalIssueID = nil
		timeEntry.PendingExternalRef = nil
		return nil
	}

	// Try to resolve the issue
	result, err := uc.ResolveIssue(ctx, userId, timeEntry.ProjectId, timeEntry.Description)
	if err != nil {
		// If resolution fails, set as pending reference
		timeEntry.ExternalIssueID = nil
		pendingRef := key
		timeEntry.PendingExternalRef = &pendingRef
		return nil // Don't fail the entire operation due to issue resolution failure
	}

	// Apply resolution result
	if result.Status == "resolved" && result.IssueID != nil {
		timeEntry.ExternalIssueID = result.IssueID
		timeEntry.PendingExternalRef = nil
	} else {
		timeEntry.ExternalIssueID = nil
		pendingRef := key
		timeEntry.PendingExternalRef = &pendingRef
	}

	return nil
}

// doesProjectBelongToUser checks if a user has read access to a project
// (either directly owns it or is a member of the team it belongs to)
func (uc *ExternalIntegrationUseCase) doesProjectBelongToUser(userID uuid.UUID, project *model.Project) bool {
	if project.UserId == userID {
		return true
	}
	if project.TeamID != nil {
		return uc.teamUsecase.DoesUserBelongToTeam(userID, *project.TeamID)
	}
	return false
}

// isUserProjectAdmin checks if a user has write access to a project
// (either directly owns it or is an admin of the team it belongs to)
func (uc *ExternalIntegrationUseCase) isUserProjectAdmin(userID uuid.UUID, project *model.Project) bool {
	if project.UserId == userID {
		return true
	}
	if project.TeamID != nil {
		return uc.teamUsecase.IsUserAdminInTeam(userID, *project.TeamID)
	}
	return false
}
