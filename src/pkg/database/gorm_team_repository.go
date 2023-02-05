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

func (repo *gormTeamRepository) GetAllTeamsOfUser(userId uuid.UUID) ([]model.Team, error) {
	// Todo: Implement in issue #22
	return nil, nil
}
