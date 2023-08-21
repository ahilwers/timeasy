package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type ProjectRepository interface {
	AddProject(project *model.Project) error
	UpdateProject(project *model.Project) error
	DeleteProject(project *model.Project) error
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetAllProjects() ([]model.Project, error)
	GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error)
}
