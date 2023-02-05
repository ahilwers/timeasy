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
	AddUserTeamAssignment(teamAssignment *model.UserTeamAssignment) error
	GetTeamsOfUser(userId uuid.UUID) ([]model.UserTeamAssignment, error)
	GetUserTeamAssignment(userId uuid.UUID, teamId uuid.UUID) (*model.UserTeamAssignment, error)
}
