package repository

import "timeasy-server/pkg/domain/model"

type ProjectRepository interface {
	AddProject(project *model.Project) (*model.Project, error)
}
