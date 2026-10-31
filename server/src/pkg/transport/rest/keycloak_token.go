package rest

import (
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v4/jwt"
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
	idStr, ok := v.token.Subject()
	if !ok || idStr == "" {
		return uuid.Nil, fmt.Errorf("token does not contain a subject")
	}
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
	realmAccess, ok := v.token.Field("realm_access")
	if !ok {
		return roles, fmt.Errorf("private claims do not contain realm_access")
	}
	realmAccessMap, ok := realmAccess.(map[string]interface{})
	if !ok {
		return roles, fmt.Errorf("realm_access is not an object")
	}
	roleIntf, ok := realmAccessMap["roles"]
	if !ok {
		return roles, fmt.Errorf("realm_access does not contain roles")
	}
	roleArray, ok := roleIntf.([]interface{})
	if !ok {
		return roles, fmt.Errorf("realm_access roles is not an array")
	}
	for _, roleIntf := range roleArray {
		role, ok := roleIntf.(string)
		if !ok {
			return roles, fmt.Errorf("realm_access contains a non-string role")
		}
		roles = append(roles, role)
	}
	return roles, nil
}
