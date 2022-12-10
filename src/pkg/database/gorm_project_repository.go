package database

import (
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

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

func (repo *gormProjectRepository) AddProject(project *model.Project) (*model.Project, error) {
	if err := repo.db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}
