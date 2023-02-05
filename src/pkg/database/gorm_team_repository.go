package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type gormTeamRepository struct {
	db *gorm.DB
}

func NewGormTeamRepository(database *gorm.DB) repository.TeamRepository {
	return &gormTeamRepository{
		db: database,
	}
}

func (repo *gormTeamRepository) AddTeam(team *model.Team) error {
	if err := repo.db.Create(team).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTeamRepository) GetTeamById(id uuid.UUID) (*model.Team, error) {
	var team model.Team
	if err := repo.db.First(&team, id).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (repo *gormTeamRepository) UpdateTeam(team *model.Team) error {
	if err := repo.db.Save(team).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTeamRepository) DeleteTeam(team *model.Team) error {
	if err := repo.db.Delete(team).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTeamRepository) GetAllTeams() ([]model.Team, error) {
	var teams []model.Team
	if err := repo.db.Order("name1").Order("name2").Order("name3").Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (repo *gormTeamRepository) AddUserTeamAssignment(teamAssignment *model.UserTeamAssignment) error {
	if err := repo.db.Create(teamAssignment).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormTeamRepository) GetTeamsOfUser(userId uuid.UUID) ([]model.UserTeamAssignment, error) {
	var assignments []model.UserTeamAssignment
	if err := repo.db.Joins("Team").Find(&assignments, "user_id=?", userId).Order("teams.name1").Order("teams.name2").Order("teams.name3").Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func (repo *gormTeamRepository) GetUserTeamAssignment(userId uuid.UUID, teamId uuid.UUID) (*model.UserTeamAssignment, error) {
	var teamAssignment model.UserTeamAssignment
	if err := repo.db.First(&teamAssignment, "user_id=? AND team_id=?", userId, teamId).Error; err != nil {
		return nil, err
	}
	return &teamAssignment, nil
}
