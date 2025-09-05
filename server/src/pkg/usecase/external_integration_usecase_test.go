package usecase

import (
	"context"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) GetProjectById(id uuid.UUID) (*model.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (m *MockProjectRepository) AddProject(project *model.Project, tx model.Transaction) error {
	args := m.Called(project, tx)
	return args.Error(0)
}

func (m *MockProjectRepository) UpdateProject(project *model.Project, tx model.Transaction) error {
	args := m.Called(project, tx)
	return args.Error(0)
}

func (m *MockProjectRepository) DeleteProject(project *model.Project, tx model.Transaction) error {
	args := m.Called(project, tx)
	return args.Error(0)
}

func (m *MockProjectRepository) GetAllProjects() ([]model.Project, error) {
	args := m.Called()
	return args.Get(0).([]model.Project), args.Error(1)
}

func (m *MockProjectRepository) GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error) {
	args := m.Called(userId)
	return args.Get(0).([]model.Project), args.Error(1)
}

func (m *MockProjectRepository) AssignProjectToTeam(project *model.Project, team *model.Team, tx model.Transaction) error {
	args := m.Called(project, team, tx)
	return args.Error(0)
}

func (m *MockProjectRepository) BeginTransaction() (model.Transaction, error) {
	args := m.Called()
	return args.Get(0).(model.Transaction), args.Error(1)
}

func createSimpleExternalIntegrationUseCase(teamUsecase TeamUsecase) (*ExternalIntegrationUseCase, *MockProjectRepository) {
	mockProjectRepo := &MockProjectRepository{}

	// Use nil for repositories that aren't needed for access control tests
	uc := NewExternalIntegrationUseCase(
		nil, // externalConnRepo
		nil, // externalIssueRepo
		nil, // userAccountRepo
		nil, // timeEntryRepo
		mockProjectRepo,
		nil, // providerFactory
		teamUsecase,
	)

	return uc, mockProjectRepo
}

func createTestProject(userID uuid.UUID, teamID *uuid.UUID) *model.Project {
	project := &model.Project{
		ID:     uuid.Must(uuid.NewV4()),
		Name:   "Test Project",
		UserId: userID,
		TeamID: teamID,
	}
	return project
}

func Test_ExternalIntegrationUseCase_doesProjectBelongToUser_DirectOwnership(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	userID := uuid.Must(uuid.NewV4())
	project := createTestProject(userID, nil)

	result := uc.doesProjectBelongToUser(userID, project)
	assert.True(t, result)
}

func Test_ExternalIntegrationUseCase_doesProjectBelongToUser_NotOwned(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	userID := uuid.Must(uuid.NewV4())
	otherUserID := uuid.Must(uuid.NewV4())
	project := createTestProject(otherUserID, nil)

	result := uc.doesProjectBelongToUser(userID, project)
	assert.False(t, result)
}

func Test_ExternalIntegrationUseCase_doesProjectBelongToUser_TeamMembership(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	projectOwnerID := GetTestUserId(t)
	teamMemberID := GetTestUserId(t)
	clientId := GetTestClientId(t)

	team := model.Team{Name1: "Test Team"}
	err := usecaseTest.TeamUsecase.AddTeam(&team, projectOwnerID, clientId)
	assert.Nil(t, err)

	_, err = usecaseTest.TeamUsecase.AddUserToTeam(teamMemberID, &team, model.RoleList{model.RoleUser}, projectOwnerID, clientId)
	assert.Nil(t, err)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	// Create project owned by project owner and assigned to team
	project := createTestProject(projectOwnerID, &team.ID)

	result := uc.doesProjectBelongToUser(teamMemberID, project)
	assert.True(t, result)

	nonMemberID := GetTestUserId(t)
	result = uc.doesProjectBelongToUser(nonMemberID, project)
	assert.False(t, result)
}

func Test_ExternalIntegrationUseCase_isUserProjectAdmin_DirectOwnership(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	userID := uuid.Must(uuid.NewV4())
	project := createTestProject(userID, nil)

	result := uc.isUserProjectAdmin(userID, project)
	assert.True(t, result)
}

func Test_ExternalIntegrationUseCase_isUserProjectAdmin_NotOwned(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	userID := uuid.Must(uuid.NewV4())
	otherUserID := uuid.Must(uuid.NewV4())
	project := createTestProject(otherUserID, nil)

	result := uc.isUserProjectAdmin(userID, project)
	assert.False(t, result)
}

func Test_ExternalIntegrationUseCase_isUserProjectAdmin_TeamAdmin(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	projectOwnerID := GetTestUserId(t)
	teamAdminID := GetTestUserId(t)
	teamMemberID := GetTestUserId(t)
	clientId := GetTestClientId(t)

	team := model.Team{Name1: "Test Team"}
	err := usecaseTest.TeamUsecase.AddTeam(&team, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Add team admin to the team with admin role
	_, err = usecaseTest.TeamUsecase.AddUserToTeam(teamAdminID, &team, model.RoleList{model.RoleUser, model.RoleAdmin}, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Add regular team member without admin role
	_, err = usecaseTest.TeamUsecase.AddUserToTeam(teamMemberID, &team, model.RoleList{model.RoleUser}, projectOwnerID, clientId)
	assert.Nil(t, err)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	// Create project owned by project owner and assigned to team
	project := createTestProject(projectOwnerID, &team.ID)

	// Test that team admin has admin access
	result := uc.isUserProjectAdmin(teamAdminID, project)
	assert.True(t, result)

	// Test that regular team member doesn't have admin access
	result = uc.isUserProjectAdmin(teamMemberID, project)
	assert.False(t, result)

	// Test that non-team member doesn't have admin access
	nonMemberID := GetTestUserId(t)
	result = uc.isUserProjectAdmin(nonMemberID, project)
	assert.False(t, result)
}

func Test_ExternalIntegrationUseCase_ConnectProjectToAccount_AccessDeniedForNonOwner(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, mockProjectRepo := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	ctx := context.Background()
	userID := uuid.Must(uuid.NewV4())
	otherUserID := uuid.Must(uuid.NewV4())
	projectID := uuid.Must(uuid.NewV4())
	userAccountID := uuid.Must(uuid.NewV4())

	project := createTestProject(otherUserID, nil)
	mockProjectRepo.On("GetProjectById", projectID).Return(project, nil)

	err := uc.ConnectProjectToAccount(ctx, userID, projectID, userAccountID, "test/repo")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "access denied")

	mockProjectRepo.AssertExpectations(t)
}

func Test_ExternalIntegrationUseCase_GetDescriptionSuggestions_AccessDeniedForNonMember(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, mockProjectRepo := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	ctx := context.Background()
	userID := uuid.Must(uuid.NewV4())
	otherUserID := uuid.Must(uuid.NewV4())
	projectID := uuid.Must(uuid.NewV4())

	project := createTestProject(otherUserID, nil)
	mockProjectRepo.On("GetProjectById", projectID).Return(project, nil)

	_, err := uc.GetDescriptionSuggestions(ctx, userID, projectID, "query", 10)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "access denied")

	mockProjectRepo.AssertExpectations(t)
}

func Test_ExternalIntegrationUseCase_DetectIssuePattern_GitHub(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	tests := []struct {
		name             string
		input            string
		expectedKey      string
		expectedProvider string
	}{
		{
			name:             "GitHub issue number simple",
			input:            "Fix issue #123",
			expectedKey:      "#123",
			expectedProvider: "github",
		},
		{
			name:             "GitHub issue number at start",
			input:            "#456 needs to be fixed",
			expectedKey:      "#456",
			expectedProvider: "github",
		},
		{
			name:             "GitHub issue number with owner/repo",
			input:            "Fixes owner/repo#789",
			expectedKey:      "#789",
			expectedProvider: "github",
		},
		{
			name:             "GitHub issue in middle of text",
			input:            "This is related to #101 and needs review",
			expectedKey:      "#101",
			expectedProvider: "github",
		},
		{
			name:             "GitHub issue with dots and dashes in repo name",
			input:            "Fixed bug my-org/my.repo#999",
			expectedKey:      "#999",
			expectedProvider: "github",
		},
		{
			name:             "No GitHub issue pattern",
			input:            "Just regular text without issue",
			expectedKey:      "",
			expectedProvider: "",
		},
		{
			name:             "Hash without number",
			input:            "Use # for comments",
			expectedKey:      "",
			expectedProvider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, provider := uc.DetectIssuePattern(tt.input)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedProvider, provider)
		})
	}
}

func Test_ExternalIntegrationUseCase_DetectIssuePattern_GitLab(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	// GitLab uses the same pattern as GitHub, so test similar cases
	tests := []struct {
		name             string
		input            string
		expectedKey      string
		expectedProvider string
	}{
		{
			name:             "GitLab issue number simple",
			input:            "Resolve issue #789",
			expectedKey:      "#789",
			expectedProvider: "github", // Note: Currently returns "github" for GitLab patterns too
		},
		{
			name:             "GitLab issue with group/project",
			input:            "Fixed in group/project#234",
			expectedKey:      "#234",
			expectedProvider: "github",
		},
		{
			name:             "GitLab merge request reference",
			input:            "Closes !567",
			expectedKey:      "",
			expectedProvider: "", // MR patterns not supported yet
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, provider := uc.DetectIssuePattern(tt.input)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedProvider, provider)
		})
	}
}

func Test_ExternalIntegrationUseCase_DetectIssuePattern_Jira(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	tests := []struct {
		name             string
		input            string
		expectedKey      string
		expectedProvider string
	}{
		{
			name:             "Jira issue simple",
			input:            "Fix bug PROJ-123",
			expectedKey:      "PROJ-123",
			expectedProvider: "jira",
		},
		{
			name:             "Jira issue at start",
			input:            "ABC-456 is broken",
			expectedKey:      "ABC-456",
			expectedProvider: "jira",
		},
		{
			name:             "Jira issue in middle",
			input:            "This relates to TICKET-789 and needs work",
			expectedKey:      "TICKET-789",
			expectedProvider: "jira",
		},
		{
			name:             "Jira issue with numbers in project key",
			input:            "Fixed PROJ2024-999",
			expectedKey:      "PROJ2024-999",
			expectedProvider: "jira",
		},
		{
			name:             "Multiple Jira issues - first one detected",
			input:            "PROJ-123 and PROJ-456 are related",
			expectedKey:      "PROJ-123",
			expectedProvider: "jira",
		},
		{
			name:             "Invalid Jira format - lowercase",
			input:            "proj-123 is not valid",
			expectedKey:      "",
			expectedProvider: "",
		},
		{
			name:             "Invalid Jira format - no dash",
			input:            "PROJ123 is not valid",
			expectedKey:      "",
			expectedProvider: "",
		},
		{
			name:             "Invalid Jira format - starts with number",
			input:            "123-PROJ is not valid",
			expectedKey:      "",
			expectedProvider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, provider := uc.DetectIssuePattern(tt.input)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedProvider, provider)
		})
	}
}

func Test_ExternalIntegrationUseCase_DetectIssuePattern_Mixed(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	tests := []struct {
		name             string
		input            string
		expectedKey      string
		expectedProvider string
		description      string
	}{
		{
			name:             "Both GitHub and Jira - GitHub detected first",
			input:            "Fix #123 and PROJ-456",
			expectedKey:      "#123",
			expectedProvider: "github",
			description:      "Should detect GitHub pattern first when both are present",
		},
		{
			name:             "Both Jira and GitHub - GitHub priority",
			input:            "PROJ-789 relates to #101",
			expectedKey:      "#101",
			expectedProvider: "github",
			description:      "Should detect GitHub pattern first due to priority, even when Jira appears first",
		},
		{
			name:             "Empty string",
			input:            "",
			expectedKey:      "",
			expectedProvider: "",
			description:      "Should handle empty input gracefully",
		},
		{
			name:             "Only whitespace",
			input:            "   \t\n   ",
			expectedKey:      "",
			expectedProvider: "",
			description:      "Should handle whitespace-only input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, provider := uc.DetectIssuePattern(tt.input)
			assert.Equal(t, tt.expectedKey, key, tt.description)
			assert.Equal(t, tt.expectedProvider, provider, tt.description)
		})
	}
}

func Test_ExternalIntegrationUseCase_ProcessTimeEntryIssueDetection_EmptyDescription(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	ctx := context.Background()
	userID := uuid.Must(uuid.NewV4())
	projectID := uuid.Must(uuid.NewV4())

	timeEntry := &model.TimeEntry{
		ID:          uuid.Must(uuid.NewV4()),
		UserId:      userID,
		ProjectId:   projectID,
		Description: "",
	}

	err := uc.ProcessTimeEntryIssueDetection(ctx, timeEntry, userID)
	assert.Nil(t, err)
	assert.Nil(t, timeEntry.ExternalIssueID)
	assert.Nil(t, timeEntry.PendingExternalRef)
}

func Test_ExternalIntegrationUseCase_ProcessTimeEntryIssueDetection_NoIssuePattern(t *testing.T) {
	usecaseTest := NewUsecaseTest()
	teardownTest := usecaseTest.SetupTest(t)
	defer teardownTest(t)

	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	ctx := context.Background()
	userID := uuid.Must(uuid.NewV4())
	projectID := uuid.Must(uuid.NewV4())

	// Set up a time entry with existing external data that should be cleared
	existingIssueID := uuid.Must(uuid.NewV4())
	existingPendingRef := "OLD-123"

	timeEntry := &model.TimeEntry{
		ID:                 uuid.Must(uuid.NewV4()),
		UserId:             userID,
		ProjectId:          projectID,
		Description:        "Just regular work without issue reference",
		ExternalIssueID:    &existingIssueID,
		PendingExternalRef: &existingPendingRef,
	}

	err := uc.ProcessTimeEntryIssueDetection(ctx, timeEntry, userID)
	assert.Nil(t, err)
	assert.Nil(t, timeEntry.ExternalIssueID, "Should clear existing external issue ID")
	assert.Nil(t, timeEntry.PendingExternalRef, "Should clear existing pending reference")
}
