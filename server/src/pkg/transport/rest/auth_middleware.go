package rest

import "github.com/gin-gonic/gin"

type AuthMiddleware interface {
	HandlerFunc() gin.HandlerFunc
}
