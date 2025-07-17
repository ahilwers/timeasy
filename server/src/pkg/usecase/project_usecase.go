package usecase

import (
	"errors"
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
	AssignProjectToTeam(project *model.Project, team *model.Team) error
}

type projectUsecase struct {
	repo          repository.ProjectRepository
	changelogRepo repository.ChangelogRepository
	teamUsecase   TeamUsecase
}

func NewProjectUsecase(repo repository.ProjectRepository, teamUsecase TeamUsecase, changelogRepo repository.ChangelogRepository) ProjectUsecase {
	return &projectUsecase{
		repo:          repo,
		changelogRepo: changelogRepo,
		teamUsecase:   teamUsecase,
	}
}

func (pu *projectUsecase) AddProject(project *model.Project) error {
	if project.UserId == uuid.Nil {
		return NewEntityIncompleteError("the user id must not be empty")
	}
	tx, err := pu.repo.BeginTransaction()
	if err != nil {
		return err
	}

	err = pu.repo.AddProject(project, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	changelogEntry := model.ChangelogEntry{
		EntityType:    model.EntityTypeProject,
		EntityID:      project.ID,
		Operation:     model.OperationCreated,
		ChangedByUser: project.UserId,
	}
	err = pu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
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
		return pu.getError(err)
	}
	tx, err := pu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	err = pu.repo.UpdateProject(project, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	changelogEntry := model.ChangelogEntry{
		EntityType:    model.EntityTypeProject,
		EntityID:      project.ID,
		Operation:     model.OperationUpdated,
		ChangedByUser: project.UserId,
	}
	err = pu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (pu *projectUsecase) DeleteProject(id uuid.UUID) error {
	project, err := pu.GetProjectById(id)
	if err != nil {
		return pu.getError(err)
	}
	tx, err := pu.repo.BeginTransaction()
	if err != nil {
		return err
	}
	err = pu.repo.DeleteProject(project, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	changelogEntry := model.ChangelogEntry{
		EntityType:    model.EntityTypeProject,
		EntityID:      project.ID,
		Operation:     model.OperationDeleted,
		ChangedByUser: project.UserId,
	}
	err = pu.changelogRepo.AddChangelogEntry(&changelogEntry, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (pu *projectUsecase) GetAllProjects() ([]model.Project, error) {
	return pu.repo.GetAllProjects()
}

func (pu *projectUsecase) GetAllProjectsOfUser(userId uuid.UUID) ([]model.Project, error) {
	return pu.repo.GetAllProjectsOfUser(userId)
}

func (pu *projectUsecase) AssignProjectToTeam(project *model.Project, team *model.Team) error {
	_, err := pu.GetProjectById(project.ID)
	if err != nil {
		return pu.getError(err)
	}
	_, err = pu.teamUsecase.GetTeamById(team.ID)
	if err != nil {
		return pu.getError(err)
	}
	project.TeamID = &team.ID
	err = pu.UpdateProject(project)
	return err
}

func (usecase *projectUsecase) getError(err error) error {
	if errors.Is(err, repository.ErrEntityNotFound) {
		return NewEntityNotFoundError(err.Error())
	}
	return err
}
