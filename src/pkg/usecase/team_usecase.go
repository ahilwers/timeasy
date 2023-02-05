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
	AddTeam(team *model.Team, owner *model.User) error
	UpdateTeam(team *model.Team) error
	DeleteTeam(id uuid.UUID) error
	GetTeamsOfUser(userId uuid.UUID) ([]model.UserTeamAssignment, error)
	AddUserToTeam(user *model.User, team model.Team, roles model.RoleList) (*model.UserTeamAssignment, error)
}

type teamUsecase struct {
	repo repository.TeamRepository
}

func NewTeamUsecase(repo repository.TeamRepository) TeamUsecase {
	return &teamUsecase{
		repo: repo,
	}
}

func (usecase *teamUsecase) AddTeam(team *model.Team, owner *model.User) error {
	err := usecase.repo.AddTeam(team)
	if err != nil {
		return err
	}
	_, err = usecase.AddUserToTeam(owner, *team, model.RoleList{model.RoleUser, model.RoleAdmin})
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

func (usecase *teamUsecase) AddUserToTeam(user *model.User, team model.Team, roles model.RoleList) (*model.UserTeamAssignment, error) {
	_, err := usecase.repo.GetUserTeamAssignment(user.ID, team.ID)
	// if this throws no error the assignment already exists:
	if err == nil {
		return nil, NewEntityExistsError(fmt.Sprintf("an assignment between user %v and team %v already exists", user.ID, team.ID))
	}

	assignment := model.UserTeamAssignment{
		UserID: user.ID,
		TeamID: team.ID,
		Roles:  roles,
	}

	err = usecase.repo.AddUserTeamAssignment(&assignment)
	if err != nil {
		return nil, err
	}

	return &assignment, nil
}
