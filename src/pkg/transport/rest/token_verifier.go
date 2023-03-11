package rest

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type TokenVerifier interface {
	VerifyToken(c *gin.Context) (*jwt.Token, error)
	GetUserId(token *jwt.Token) (uuid.UUID, error)
	GetRoles(token *jwt.Token) ([]string, error)
	HasRole(token *jwt.Token, role string) (bool, error)
}
