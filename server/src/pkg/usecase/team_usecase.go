package usecase

import (
	"errors"
	"fmt"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type TeamUsecase interface {
	GetTeamById(id uuid.UUID) (*model.Team, error)
	GetAllTeams() ([]model.Team, error)
	AddTeam(team *model.Team, ownerId uuid.UUID, clientId string) error
	UpdateTeam(team *model.Team, userId uuid.UUID, clientId string) error
	DeleteTeam(id uuid.UUID, userId uuid.UUID, clientId string) error
	GetTeamsOfUser(userId uuid.UUID) ([]model.UserTeamAssignment, error)
	DoesUserBelongToTeam(userId uuid.UUID, teamId uuid.UUID) bool
	AddUserToTeam(userId uuid.UUID, team *model.Team, roles model.RoleList, changedByUserId uuid.UUID, clientId string) (*model.UserTeamAssignment, error)
	DeleteUserFromTeam(userId uuid.UUID, team *model.Team, changedByUserId uuid.UUID, clientId string) error
	UpdateUserRolesInTeam(userId uuid.UUID, team *model.Team, roles model.RoleList, changedByUserId uuid.UUID, clientId string) error
	IsUserAdminInTeam(userId uuid.UUID, teamId uuid.UUID) bool
}

type teamUsecase struct {
	repo          repository.TeamRepository
	changelogRepo repository.ChangelogRepository
}

func NewTeamUsecase(repo repository.TeamRepository, changelogRepo repository.ChangelogRepository) TeamUsecase {
	return &teamUsecase{
		repo:          repo,
		changelogRepo: changelogRepo,
	}
}

func (usecase *teamUsecase) AddTeam(team *model.Team, ownerId uuid.UUID, clientId string) error {
	tx, err := usecase.repo.BeginTransaction()
	if err != nil {
		return err
	}

	err = usecase.repo.AddTeam(team)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Add changelog entry for team creation
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeTeam,
		EntityID:        team.ID,
		Operation:       model.OperationCreated,
		ChangedByUser:   ownerId,
		ChangedByClient: clientId,
	}
	err = usecase.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = usecase.AddUserToTeam(ownerId, team, model.RoleList{model.RoleUser, model.RoleAdmin}, ownerId, clientId)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (usecase *teamUsecase) GetTeamById(id uuid.UUID) (*model.Team, error) {
	team, err := usecase.repo.GetTeamById(id)
	if err != nil {
		return nil, usecase.getError(err)
	}
	return team, nil
}

func (usecase *teamUsecase) UpdateTeam(team *model.Team, userId uuid.UUID, clientId string) error {
	_, err := usecase.GetTeamById(team.ID)
	if err != nil {
		return err
	}

	tx, err := usecase.repo.BeginTransaction()
	if err != nil {
		return err
	}

	err = usecase.repo.UpdateTeam(team)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Add changelog entry for team update
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeTeam,
		EntityID:        team.ID,
		Operation:       model.OperationUpdated,
		ChangedByUser:   userId,
		ChangedByClient: clientId,
	}
	err = usecase.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (usecase *teamUsecase) DeleteTeam(id uuid.UUID, userId uuid.UUID, clientId string) error {
	team, err := usecase.GetTeamById(id)
	if err != nil {
		return err
	}
	tx, err := usecase.repo.BeginTransaction()
	if err != nil {
		return err
	}
	err = usecase.repo.DeleteAllUserAssignmentsOfTeam(team.ID, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = usecase.repo.DeleteTeam(team, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Add changelog entry for team deletion
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeTeam,
		EntityID:        team.ID,
		Operation:       model.OperationDeleted,
		ChangedByUser:   userId,
		ChangedByClient: clientId,
	}
	err = usecase.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (usecase *teamUsecase) GetAllTeams() ([]model.Team, error) {
	return usecase.repo.GetAllTeams()
}

func (usecase *teamUsecase) GetTeamsOfUser(userId uuid.UUID) ([]model.UserTeamAssignment, error) {
	return usecase.repo.GetTeamsOfUser(userId)
}

func (usecase *teamUsecase) DoesUserBelongToTeam(userId uuid.UUID, teamId uuid.UUID) bool {
	teamAssignments, err := usecase.GetTeamsOfUser(userId)
	if err != nil {
		return false
	}
	for _, teamAssignment := range teamAssignments {
		if teamAssignment.TeamID == teamId {
			return true
		}
	}
	return false
}

func (usecase *teamUsecase) AddUserToTeam(userId uuid.UUID, team *model.Team, roles model.RoleList, changedByUserId uuid.UUID, clientId string) (*model.UserTeamAssignment, error) {
	existingAssignment, err := usecase.repo.GetUserTeamAssignment(userId, team.ID)
	if existingAssignment != nil {
		return nil, NewEntityExistsError(fmt.Sprintf("user %v already belongs to team %v", userId, team.ID))
	}

	if len(roles) == 0 {
		roles = append(roles, model.RoleUser)
	}

	assignment := model.UserTeamAssignment{
		UserID: userId,
		TeamID: team.ID,
		Roles:  roles,
	}

	tx, err := usecase.repo.BeginTransaction()
	if err != nil {
		return nil, err
	}

	err = usecase.repo.AddUserTeamAssignment(&assignment)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Add changelog entry for user team assignment
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeUserTeamAssignment,
		EntityID:        assignment.ID,
		Operation:       model.OperationCreated,
		ChangedByUser:   changedByUserId,
		ChangedByClient: clientId,
	}
	err = usecase.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &assignment, nil
}

func (usecase *teamUsecase) DeleteUserFromTeam(userId uuid.UUID, team *model.Team, changedByUserId uuid.UUID, clientId string) error {
	teamAssignment, err := usecase.repo.GetUserTeamAssignment(userId, team.ID)
	if teamAssignment == nil {
		return NewEntityNotFoundError(fmt.Sprintf("user %v does not belong to team %v", userId, team.ID))
	}

	tx, err := usecase.repo.BeginTransaction()
	if err != nil {
		return err
	}

	err = usecase.repo.DeleteUserTeamAssignment(teamAssignment)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Add changelog entry for user team assignment deletion
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeUserTeamAssignment,
		EntityID:        teamAssignment.ID,
		Operation:       model.OperationDeleted,
		ChangedByUser:   changedByUserId,
		ChangedByClient: clientId,
	}
	err = usecase.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (usecase *teamUsecase) UpdateUserRolesInTeam(userId uuid.UUID, team *model.Team, roles model.RoleList, changedByUserId uuid.UUID, clientId string) error {
	teamAssignment, err := usecase.repo.GetUserTeamAssignment(userId, team.ID)
	if teamAssignment == nil {
		return NewEntityNotFoundError(fmt.Sprintf("user %v does not belong to team %v", userId, team.ID))
	}

	tx, err := usecase.repo.BeginTransaction()
	if err != nil {
		return err
	}

	teamAssignment.Roles = roles
	err = usecase.repo.UpdateUserTeamAssignment(teamAssignment)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Add changelog entry for user team assignment update
	changelogEntry := model.ChangelogEntry{
		EntityType:      model.EntityTypeUserTeamAssignment,
		EntityID:        teamAssignment.ID,
		Operation:       model.OperationUpdated,
		ChangedByUser:   changedByUserId,
		ChangedByClient: clientId,
	}
	err = usecase.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (usecase *teamUsecase) IsUserAdminInTeam(userId uuid.UUID, teamId uuid.UUID) bool {
	teamAssignment, _ := usecase.repo.GetUserTeamAssignment(userId, teamId)
	if teamAssignment == nil {
		return false
	}
	return usecase.hasRole(teamAssignment.Roles, model.RoleAdmin)
}

func (usecase *teamUsecase) hasRole(roles model.RoleList, role string) bool {
	for _, role := range roles {
		if role == model.RoleAdmin {
			return true
		}
	}
	return false
}

func (usecase *teamUsecase) getError(err error) error {
	if errors.Is(err, repository.ErrEntityNotFound) {
		return NewEntityNotFoundError(err.Error())
	}
	return err
}
