package repository

import (
	"timeasy-server/pkg/domain/model"

	"github.com/gofrs/uuid"
)

type ProjectRepository interface {
	model.TransactionHandler
	AddProject(project *model.Project, tx model.Transaction) error
	UpdateProject(project *model.Project, tx model.Transaction) error
	DeleteProject(project *model.Project, tx model.Transaction) error
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetAllProjects() ([]model.Project, error)
	GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error)
}
