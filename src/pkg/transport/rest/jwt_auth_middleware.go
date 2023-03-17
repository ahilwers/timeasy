package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
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
			glog.Errorf("error verifying token: %v", err)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}
