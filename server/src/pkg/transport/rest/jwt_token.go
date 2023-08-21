package rest

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
)

type jwtToken struct {
	token jwt.Token
}

func NewJwtToken(token jwt.Token) AuthToken {
	return &jwtToken{
		token: token,
	}
}

func (v *jwtToken) GetUserId() (uuid.UUID, error) {
	claims, ok := v.token.Claims.(jwt.MapClaims)
	if ok && v.token.Valid {
		uid, err := uuid.FromString(fmt.Sprintf("%v", claims["user_id"]))
		if err != nil {
			return uuid.Nil, err
		}
		return uid, nil
	}
	return uuid.Nil, fmt.Errorf("user id could not be extracted from token.")
}

func (v *jwtToken) HasRole(role string) (bool, error) {
	roles, err := v.GetRoles()
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

func (v *jwtToken) GetRoles() ([]string, error) {
	claims, ok := v.token.Claims.(jwt.MapClaims)
	if ok && v.token.Valid {
		roleString := fmt.Sprintf("%v", claims["user_roles"])
		if roleString == "" {
			return nil, fmt.Errorf("could not extract user roles from token")
		}
		return strings.Split(roleString, ","), nil
	}
	return nil, fmt.Errorf("user roles could not be extracted from token.")
}
