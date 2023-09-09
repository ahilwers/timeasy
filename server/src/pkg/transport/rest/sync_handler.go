package rest

import (
	"github.com/gin-gonic/gin"
)

type SyncHandler interface {
	GetChangedTimeEntries(context *gin.Context)
	SendLocallyChangedTimeEntries(context *gin.Context)
	GetChangedProjects(context *gin.Context)
	SendLocallyChangedProjects(context *gin.Context)
}

type syncHandler struct {
	tokenVerifier TokenVerifier
}

func NewSyncHandler(tokenVerifier TokenVerifier) SyncHandler {
	return &syncHandler{
		tokenVerifier: tokenVerifier,
	}
}

func (handler *syncHandler) GetChangedTimeEntries(context *gin.Context) {
}

func (handler *syncHandler) SendLocallyChangedTimeEntries(context *gin.Context) {
}

func (handler *syncHandler) GetChangedProjects(context *gin.Context) {
}

func (handler *syncHandler) SendLocallyChangedProjects(context *gin.Context) {
}
