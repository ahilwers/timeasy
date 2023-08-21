package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type gormProjectRepository struct {
	db             *gorm.DB
	teamRepository repository.TeamRepository
}

func NewGormProjectRepository(database *gorm.DB, teamRepository repository.TeamRepository) repository.ProjectRepository {
	return &gormProjectRepository{
		db:             database,
		teamRepository: teamRepository,
	}
}

func (repo *gormProjectRepository) AddProject(project *model.Project) error {
	if err := repo.db.Create(project).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormProjectRepository) GetProjectById(id uuid.UUID) (*model.Project, error) {
	var project model.Project
	if err := repo.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (repo *gormProjectRepository) UpdateProject(project *model.Project) error {
	if err := repo.db.Save(project).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormProjectRepository) DeleteProject(project *model.Project) error {
	if err := repo.db.Delete(project).Error; err != nil {
		return err
	}
	return nil
}

func (repo *gormProjectRepository) GetAllProjects() ([]model.Project, error) {
	var projects []model.Project
	if err := repo.db.Order("name").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (repo *gormProjectRepository) GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error) {
	var projects []model.Project
	query := repo.db.Order("name")
	teamIds, err := repo.getTeamIdsOfUser(userId)
	if err != nil {
		return projects, err
	}
	if len(teamIds) != 0 {
		query = query.Where("user_id=? OR team_id IN ?", userId, teamIds)
	} else {
		query = query.Where("user_id=?", userId)
	}

	if err := query.Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (repo *gormProjectRepository) getTeamIdsOfUser(userId uuid.UUID) ([]uuid.UUID, error) {
	var teamIds []uuid.UUID
	teamAssignments, err := repo.teamRepository.GetTeamsOfUser(userId)
	if err != nil {
		return teamIds, err
	}
	for _, teamAssignment := range teamAssignments {
		teamIds = append(teamIds, teamAssignment.TeamID)
	}
	return teamIds, nil
}
