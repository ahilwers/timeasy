package rest

import (
	"strings"

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
	roles, err := v.GetRoles()
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if strings.EqualFold(r, role) {
			return true, nil
		}
	}
	return false, nil
}

func (v *keycloakToken) GetRoles() ([]string, error) {
	roles := []string{}
	realmAccess := v.token.PrivateClaims()["realm_access"]
	roleArray := realmAccess.(map[string]interface{})["roles"].([]interface{})
	for _, roleIntf := range roleArray {
		roles = append(roles, roleIntf.(string))
	}
	return roles, nil
}
