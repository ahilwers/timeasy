package usecase

import (
	"fmt"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type TeamUsecase interface {
	GetTeamById(id uuid.UUID) (*model.Team, error)
	GetAllTeams() ([]model.Team, error)
	AddTeam(team *model.Team, ownerId uuid.UUID) error
	UpdateTeam(team *model.Team) error
	DeleteTeam(id uuid.UUID) error
	GetTeamsOfUser(userId uuid.UUID) ([]model.UserTeamAssignment, error)
	DoesUserBelongToTeam(userId uuid.UUID, teamId uuid.UUID) bool
	AddUserToTeam(userId uuid.UUID, team *model.Team, roles model.RoleList) (*model.UserTeamAssignment, error)
	DeleteUserFromTeam(userId uuid.UUID, team *model.Team) error
	UpdateUserRolesInTeam(userId uuid.UUID, team *model.Team, roles model.RoleList) error
	IsUserAdminInTeam(userId uuid.UUID, teamId uuid.UUID) bool
}

type teamUsecase struct {
	repo repository.TeamRepository
}

func NewTeamUsecase(repo repository.TeamRepository) TeamUsecase {
	return &teamUsecase{
		repo: repo,
	}
}

func (usecase *teamUsecase) AddTeam(team *model.Team, ownerId uuid.UUID) error {
	err := usecase.repo.AddTeam(team)
	if err != nil {
		return err
	}
	_, err = usecase.AddUserToTeam(ownerId, team, model.RoleList{model.RoleUser, model.RoleAdmin})
	if err != nil {
		return err
	}
	return nil
}

func (usecase *teamUsecase) GetTeamById(id uuid.UUID) (*model.Team, error) {
	team, err := usecase.repo.GetTeamById(id)
	if err != nil {
		return nil, NewEntityNotFoundError(fmt.Sprintf("team with id %v not found", id))
	}
	return team, nil
}

func (usecase *teamUsecase) UpdateTeam(team *model.Team) error {
	_, err := usecase.GetTeamById(team.ID)
	if err != nil {
		return err
	}
	return usecase.repo.UpdateTeam(team)
}

func (usecase *teamUsecase) DeleteTeam(id uuid.UUID) error {
	team, err := usecase.GetTeamById(id)
	if err != nil {
		return err
	}
	return usecase.repo.DeleteTeam(team)
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

func (usecase *teamUsecase) AddUserToTeam(userId uuid.UUID, team *model.Team, roles model.RoleList) (*model.UserTeamAssignment, error) {
	_, err := usecase.repo.GetUserTeamAssignment(userId, team.ID)
	// if this throws no error the assignment already exists:
	if err == nil {
		return nil, NewEntityExistsError(fmt.Sprintf("an assignment between user %v and team %v already exists", userId, team.ID))
	}

	// If no roles a re given add the user role:
	if len(roles) == 0 {
		roles = append(roles, model.RoleUser)
	}

	assignment := model.UserTeamAssignment{
		UserID: userId,
		TeamID: team.ID,
		Roles:  roles,
	}

	err = usecase.repo.AddUserTeamAssignment(&assignment)
	if err != nil {
		return nil, err
	}

	return &assignment, nil
}

func (usecase *teamUsecase) DeleteUserFromTeam(userId uuid.UUID, team *model.Team) error {
	teamAssignment, err := usecase.repo.GetUserTeamAssignment(userId, team.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("assignment between user %v and team %v not found", userId, team.ID))
	}
	err = usecase.repo.DeleteUserTeamAssignment(teamAssignment)
	if err != nil {
		return err
	}
	return nil
}

func (usecase *teamUsecase) UpdateUserRolesInTeam(userId uuid.UUID, team *model.Team, roles model.RoleList) error {
	teamAssignment, err := usecase.repo.GetUserTeamAssignment(userId, team.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("assignment between user %v and team %v not found", userId, team.ID))
	}
	teamAssignment.Roles = roles
	err = usecase.repo.UpdateUserTeamAssignment(teamAssignment)
	if err != nil {
		return err
	}
	return nil
}

func (usecase *teamUsecase) IsUserAdminInTeam(userId uuid.UUID, teamId uuid.UUID) bool {
	teamAssignment, err := usecase.repo.GetUserTeamAssignment(userId, teamId)
	if err != nil {
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
