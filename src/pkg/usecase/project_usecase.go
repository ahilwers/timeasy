package usecase

import (
	"fmt"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type ProjectUsecase interface {
	GetProjectById(id uuid.UUID) (*model.Project, error)
	GetAllProjects() ([]model.Project, error)
	GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error)
	AddProject(project *model.Project) error
	UpdateProject(project *model.Project) error
	DeleteProject(id uuid.UUID) error
}

type projectUsecase struct {
	repo repository.ProjectRepository
}

func NewProjectUsecase(repo repository.ProjectRepository) ProjectUsecase {
	return &projectUsecase{
		repo: repo,
	}
}

func (pu *projectUsecase) AddProject(project *model.Project) error {
	if project.UserId == uuid.Nil {
		return NewEntityIncompleteError("the user id must not be empty")
	}
	return pu.repo.AddProject(project)
}

func (pu *projectUsecase) GetProjectById(id uuid.UUID) (*model.Project, error) {
	return pu.repo.GetProjectById(id)
}

func (pu *projectUsecase) UpdateProject(project *model.Project) error {
	if project.UserId == uuid.Nil {
		return NewEntityIncompleteError("the user id must not be empty")
	}
	_, err := pu.GetProjectById(project.ID)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("project with id %v does not exist", project.ID))
	}
	return pu.repo.UpdateProject(project)
}

func (pu *projectUsecase) DeleteProject(id uuid.UUID) error {
	project, err := pu.GetProjectById(id)
	if err != nil {
		return NewEntityNotFoundError(fmt.Sprintf("project with id %v does not exist", id))
	}
	return pu.repo.DeleteProject(project)
}

func (pu *projectUsecase) GetAllProjects() ([]model.Project, error) {
	return pu.repo.GetAllProjects()
}

func (pu *projectUsecase) GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error) {
	return pu.repo.GetAllProjectsOfUser(userId)
}
