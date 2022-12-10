package usecase

import (
	"fmt"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"
)

type ProjectUsecase interface {
	AddProject(project *model.Project) (*model.Project, error)
}

type projectUsecase struct {
	repo repository.ProjectRepository
}

func NewProjectUsecase(repo repository.ProjectRepository) ProjectUsecase {
	return &projectUsecase{
		repo: repo,
	}
}

func (pu *projectUsecase) AddProject(project *model.Project) (*model.Project, error) {
	if project.UserId == "" {
		fmt.Println("Id not set")
		//	return nil, fmt.Errorf("The user id must not be empty.")
	}
	return pu.repo.AddProject(project)
}
