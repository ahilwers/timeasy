package repository

import "errors"

// Common repository errors
var (
	// ErrEntityNotFound is returned when an entity is not found in the repository
	ErrEntityNotFound = errors.New("entity not found")
)
