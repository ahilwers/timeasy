package model

import (
	"github.com/gofrs/uuid"
)

type Team struct {
	ID    uuid.UUID
	Name1 string
	Name2 string
	Name3 string
}
