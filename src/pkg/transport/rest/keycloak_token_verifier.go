package rest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

type keycloakTokenVerifier struct {
	keycloakHost string
	realm        string
}

func NewKeycloakTokenVerifier(keycloakHost string, realm string) TokenVerifier {
	return &keycloakTokenVerifier{
		keycloakHost: keycloakHost,
		realm:        realm,
	}
}

func (v *keycloakTokenVerifier) VerifyToken(c *gin.Context) (AuthToken, error) {
	strToken, err := v.getAuthHeader(c)
	if err != nil {
		return nil, err
	}
	jwksKeySet, err := jwk.Fetch(c.Request.Context(), v.getKeycloakJwksUrl())
	if err != nil {
		return nil, err
	}
	token, err := jwt.Parse([]byte(strToken), jwt.WithKeySet(jwksKeySet), jwt.WithValidate(true))
	if err != nil {
		return nil, err
	}
	return NewKeycloakToken(token), nil
}

func (v *keycloakTokenVerifier) getAuthHeader(c *gin.Context) (string, error) {
	header := strings.Fields(c.Request.Header.Get("Authorization"))
	if header[0] != "Bearer" {
		return "", errors.New("malformed token")
	}
	return header[1], nil
}

func (v *keycloakTokenVerifier) getKeycloakJwksUrl() string {
	url := fmt.Sprintf("%v/realms/%v/protocol/openid-connect/certs", v.keycloakHost, v.realm)
	return url
}
