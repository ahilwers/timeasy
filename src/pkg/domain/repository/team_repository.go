package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type TeamRepository interface {
	AddTeam(team *model.Team) error
	UpdateTeam(team *model.Team) error
	DeleteTeam(team *model.Team) error
	GetTeamById(id uuid.UUID) (*model.Team, error)
	GetAllTeams() ([]model.Team, error)
	GetAllTeamsOfUser(userId uuid.UUID) ([]model.Team, error)
}
