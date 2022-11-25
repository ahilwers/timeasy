package project

import "gorm.io/gorm"

func AddProject(database *gorm.DB, project *Project) (*Project, error) {
	if err := database.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}
