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
	GetAllTeamsOfUser(userId uuid.UUID) ([]model.Team, error)
	AddTeam(team *model.Team) error
	UpdateTeam(team *model.Team) error
	DeleteTeam(id uuid.UUID) error
}

type teamUsecase struct {
	repo repository.TeamRepository
}

func NewTeamUsecase(repo repository.TeamRepository) TeamUsecase {
	return &teamUsecase{
		repo: repo,
	}
}

func (pu *teamUsecase) AddTeam(team *model.Team) error {
	return pu.repo.AddTeam(team)
}

func (pu *teamUsecase) GetTeamById(id uuid.UUID) (*model.Team, error) {
	return pu.repo.GetTeamById(id)
}

func (pu *teamUsecase) UpdateTeam(team *model.Team) error {
	_, err := pu.GetTeamById(team.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("team with id %v does not exist", team.ID))
	}
	return pu.repo.UpdateTeam(team)
}

func (pu *teamUsecase) DeleteTeam(id uuid.UUID) error {
	team, err := pu.GetTeamById(id)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("team with id %v does not exist", id))
	}
	return pu.repo.DeleteTeam(team)
}

func (pu *teamUsecase) GetAllTeams() ([]model.Team, error) {
	return pu.repo.GetAllTeams()
}

func (pu *teamUsecase) GetAllTeamsOfUser(userId uuid.UUID) ([]model.Team, error) {
	return pu.repo.GetAllTeamsOfUser(userId)
}
