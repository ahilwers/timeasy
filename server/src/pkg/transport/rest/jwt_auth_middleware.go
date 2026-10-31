package rest

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

const AuthTokenContextKey = "authToken"

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
		authToken, err := mw.tokenVerifier.VerifyToken(c)
		if err != nil {
			slog.Error("Error verifying token", "error", err, "path", c.Request.URL.Path, "method", c.Request.Method)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		// Store the verified token in context for use by subsequent middleware and handlers
		c.Set(AuthTokenContextKey, authToken)
		c.Next()
	}
}

// GetAuthTokenFromContext retrieves the verified auth token from the gin context
func GetAuthTokenFromContext(c *gin.Context) (AuthToken, bool) {
	token, exists := c.Get(AuthTokenContextKey)
	if !exists {
		return nil, false
	}
	authToken, ok := token.(AuthToken)
	return authToken, ok
}
