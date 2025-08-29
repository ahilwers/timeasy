package usecase

import (
	"context"
	"testing"
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Minimal mock project repository for testing access control logic
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

// createSimpleExternalIntegrationUseCase creates a test instance with minimal mocked dependencies
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

// Test helper functions
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

	// Create users and team using real usecase (not mocked)
	projectOwnerID := GetTestUserId(t)
	teamMemberID := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create a team
	team := model.Team{Name1: "Test Team"}
	err := usecaseTest.TeamUsecase.AddTeam(&team, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Add team member to the team
	_, err = usecaseTest.TeamUsecase.AddUserToTeam(teamMemberID, &team, model.RoleList{model.RoleUser}, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Create external integration usecase with real team usecase
	uc, _ := createSimpleExternalIntegrationUseCase(usecaseTest.TeamUsecase)

	// Create project owned by project owner and assigned to team
	project := createTestProject(projectOwnerID, &team.ID)

	// Test that team member has access
	result := uc.doesProjectBelongToUser(teamMemberID, project)
	assert.True(t, result)

	// Test that non-team member doesn't have access
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

	// Create users and team using real usecase (not mocked)
	projectOwnerID := GetTestUserId(t)
	teamAdminID := GetTestUserId(t)
	teamMemberID := GetTestUserId(t)
	clientId := GetTestClientId(t)

	// Create a team
	team := model.Team{Name1: "Test Team"}
	err := usecaseTest.TeamUsecase.AddTeam(&team, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Add team admin to the team with admin role
	_, err = usecaseTest.TeamUsecase.AddUserToTeam(teamAdminID, &team, model.RoleList{model.RoleUser, model.RoleAdmin}, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Add regular team member without admin role
	_, err = usecaseTest.TeamUsecase.AddUserToTeam(teamMemberID, &team, model.RoleList{model.RoleUser}, projectOwnerID, clientId)
	assert.Nil(t, err)

	// Create external integration usecase with real team usecase
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

	// Mock project owned by different user
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

	// Mock project owned by different user with no team
	project := createTestProject(otherUserID, nil)
	mockProjectRepo.On("GetProjectById", projectID).Return(project, nil)

	_, err := uc.GetDescriptionSuggestions(ctx, userID, projectID, "query", 10)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "access denied")

	mockProjectRepo.AssertExpectations(t)
}