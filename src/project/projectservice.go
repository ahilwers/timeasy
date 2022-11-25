package project

import (
	"fmt"

	"gorm.io/gorm"
)

type ProjectService interface {
	AddProject(project *Project) (*Project, error)
}

type projectService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) ProjectService {
	return &projectService{db}
}

func (projectService *projectService) AddProject(project *Project) (*Project, error) {
	if project.UserId == "" {
		return nil, fmt.Errorf("The user id must not be empty.")
	}

	if err := projectService.db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}
