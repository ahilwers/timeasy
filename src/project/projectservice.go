package project

import (
	"fmt"

	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func (projectService *ProjectService) Init(db *gorm.DB) {
	projectService.db = db
}

func (projectService *ProjectService) AddProject(project *Project) (*Project, error) {
	if project.UserId == "" {
		return nil, fmt.Errorf("The user id must not be empty.")
	}

	if err := projectService.db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}
