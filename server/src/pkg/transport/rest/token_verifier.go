package rest

import (
	"github.com/gin-gonic/gin"
)

type TokenVerifier interface {
	VerifyToken(c *gin.Context) (AuthToken, error)
}
