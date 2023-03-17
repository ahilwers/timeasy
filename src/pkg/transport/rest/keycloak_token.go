package rest

import (
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/jwt"
)

type keycloakToken struct {
	token jwt.Token
}

func NewKeycloakToken(token jwt.Token) AuthToken {
	return &keycloakToken{
		token: token,
	}
}

func (v *keycloakToken) GetUserId() (uuid.UUID, error) {
	idStr := v.token.Subject()
	return uuid.FromString(idStr)
}

func (v *keycloakToken) HasRole(role string) (bool, error) {
	return false, nil
}

func (v *keycloakToken) GetRoles() ([]string, error) {
	return []string{}, nil
}
