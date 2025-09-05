package rest

import (
	"net/http"
	"strconv"
	"strings"

	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type ExternalIntegrationHandler interface {
	ConnectProjectToAccount(context *gin.Context)
	DisconnectProject(context *gin.Context)
	SyncProjectIssues(context *gin.Context)
	GetDescriptionSuggestions(context *gin.Context)
	ResolveIssue(context *gin.Context)
	ResolvePendingReferences(context *gin.Context)
	GetExternalIssue(context *gin.Context)
}

type externalIntegrationHandler struct {
	tokenVerifier TokenVerifier
	usecase       *usecase.ExternalIntegrationUseCase
}

type ConnectProjectToAccountRequest struct {
	UserAccountID string `json:"userAccountId" binding:"required"`
	ProjectRef    string `json:"projectRef" binding:"required"`
}

type DescriptionSuggestionsResponse struct {
	Suggestions []string `json:"suggestions"`
}

func NewExternalIntegrationHandler(tokenVerifier TokenVerifier, usecase *usecase.ExternalIntegrationUseCase) ExternalIntegrationHandler {
	return &externalIntegrationHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
	}
}

func (h *externalIntegrationHandler) ConnectProjectToAccount(c *gin.Context) {
	token, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := token.GetUserId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.FromString(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req ConnectProjectToAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userAccountID, err := uuid.FromString(req.UserAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user account ID"})
		return
	}

	if err := h.usecase.ConnectProjectToAccount(c.Request.Context(), userUUID, projectID, userAccountID, req.ProjectRef); err != nil {
		// Provide more specific error status codes based on error type
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "invalid project reference") ||
			strings.Contains(errorMsg, "project not accessible") ||
			strings.Contains(errorMsg, "authentication failed") ||
			strings.Contains(errorMsg, "invalid token") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
			return
		}

		if strings.Contains(errorMsg, "access denied") ||
			strings.Contains(errorMsg, "not found") {
			c.JSON(http.StatusForbidden, gin.H{"error": errorMsg})
			return
		}

		if strings.Contains(errorMsg, "provider not configured") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": errorMsg})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project connected successfully"})
}

func (h *externalIntegrationHandler) DisconnectProject(c *gin.Context) {
	token, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := token.GetUserId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.FromString(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.usecase.DisconnectProject(c.Request.Context(), userUUID, projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project disconnected successfully"})
}

// GetDescriptionSuggestions returns autocomplete suggestions for time entry descriptions
func (h *externalIntegrationHandler) GetDescriptionSuggestions(c *gin.Context) {
	token, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := token.GetUserId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.FromString(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	suggestions, err := h.usecase.GetDescriptionSuggestions(c.Request.Context(), userUUID, projectID, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DescriptionSuggestionsResponse{
		Suggestions: suggestions,
	})
}

// ResolveIssue attempts to resolve an issue reference
func (h *externalIntegrationHandler) ResolveIssue(c *gin.Context) {
	token, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := token.GetUserId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := uuid.FromString(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	input := c.Query("input")
	if input == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'input' is required"})
		return
	}

	result, err := h.usecase.ResolveIssue(c.Request.Context(), userUUID, projectID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.Status == "pending" {
		c.JSON(http.StatusAccepted, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SyncProjectIssues manually triggers a sync for project issues
func (h *externalIntegrationHandler) SyncProjectIssues(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.FromString(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	err = h.usecase.SyncProjectIssues(c.Request.Context(), projectID)
	if err != nil {
		errorMsg := err.Error()

		if strings.Contains(errorMsg, "authentication failed") ||
			strings.Contains(errorMsg, "invalid token") ||
			strings.Contains(errorMsg, "project not accessible") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
			return
		}

		if strings.Contains(errorMsg, "no external connection found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "No external integration configured for this project"})
			return
		}

		if strings.Contains(errorMsg, "provider not configured") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": errorMsg})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync completed successfully"})
}

func (h *externalIntegrationHandler) ResolvePendingReferences(c *gin.Context) {
	token, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userUUID, err := token.GetUserId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user ID from token"})
		return
	}

	if err := h.usecase.ResolvePendingReferences(c.Request.Context(), userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pending references resolved"})
}

func (h *externalIntegrationHandler) GetExternalIssue(c *gin.Context) {
	_, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	issueIDStr := c.Param("id")
	issueID, err := uuid.FromString(issueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue ID"})
		return
	}

	externalIssue, err := h.usecase.GetExternalIssueByID(c.Request.Context(), issueID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "External issue not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, externalIssue)
}
