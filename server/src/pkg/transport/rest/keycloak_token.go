package rest

import (
	"fmt"
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
	realmAccess, ok := v.token.PrivateClaims()["realm_access"]
	if !ok {
		return roles, fmt.Errorf("private claims do not contain realm_access")
	}
	roleIntf, ok := realmAccess.(map[string]interface{})["roles"]
	if !ok {
		return roles, fmt.Errorf("realm_access does not contain roles")
	}
	roleArray := roleIntf.([]interface{})
	for _, roleIntf := range roleArray {
		roles = append(roles, roleIntf.(string))
	}
	return roles, nil
}
