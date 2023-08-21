package usecase

import (
	"fmt"

	"github.com/gofrs/uuid"
)

type EntityExistsError struct {
	Msg string
}

func (e *EntityExistsError) Error() string {
	return e.Msg
}

func NewEntityExistsError(msg string) *EntityExistsError {
	return &EntityExistsError{
		Msg: msg,
	}
}

type EntityNotFoundError struct {
	Msg string
}

func (e *EntityNotFoundError) Error() string {
	return e.Msg
}

func NewEntityNotFoundError(msg string) *EntityNotFoundError {
	return &EntityNotFoundError{
		Msg: msg,
	}
}

type UserNotFoundError struct {
	Msg string
}

func (e *UserNotFoundError) Error() string {
	return e.Msg
}

func NewUserNotFoundError(userId uuid.UUID) *UserNotFoundError {
	return &UserNotFoundError{
		Msg: fmt.Sprintf("user with id %v not found", userId),
	}
}

type ProjectNotFoundError struct {
	Msg string
}

func (e *ProjectNotFoundError) Error() string {
	return e.Msg
}

func NewProjectNotFoundError(projectId uuid.UUID) *ProjectNotFoundError {
	return &ProjectNotFoundError{
		Msg: fmt.Sprintf("project with id %v not found", projectId),
	}
}

type EntityIncompleteError struct {
	Msg string
}

func (e *EntityIncompleteError) Error() string {
	return e.Msg
}

func NewEntityIncompleteError(msg string) *EntityIncompleteError {
	return &EntityIncompleteError{
		Msg: msg,
	}
}
