package rest

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type jwtAuthMiddleware struct {
	tokenVerifier TokenVerifier
}

func NewJwtAuthMiddleware(tokenVerifier TokenVerifier) AuthMiddleware {
	return &jwtAuthMiddleware{
		tokenVerifier: tokenVerifier,
	}
}

func (mw *jwtAuthMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := mw.tokenVerifier.VerifyToken(c)
		if err != nil {
			slog.Error("Error verifying token", "error", err, "path", c.Request.URL.Path, "method", c.Request.Method)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}
