package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type gormProjectRepository struct {
	db *gorm.DB
}

func NewGormProjectRepository(database *gorm.DB) repository.ProjectRepository {
	return &gormProjectRepository{
		db: database,
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
	if err := repo.db.Order("name").Find(&projects, "user_id=?", userId).Error; err != nil {
		return nil, err
	}
	return projects, nil
}
