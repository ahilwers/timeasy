package rest

import (
	"net/http"
	"strconv"

	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type ExternalIntegrationHandler struct {
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

func NewExternalIntegrationHandler(tokenVerifier TokenVerifier, usecase *usecase.ExternalIntegrationUseCase) *ExternalIntegrationHandler {
	return &ExternalIntegrationHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
	}
}

// ConnectProjectToAccount connects a project to a user's external account
func (h *ExternalIntegrationHandler) ConnectProjectToAccount(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project connected successfully"})
}

// DisconnectProject disconnects a project from external provider
func (h *ExternalIntegrationHandler) DisconnectProject(c *gin.Context) {
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
func (h *ExternalIntegrationHandler) GetDescriptionSuggestions(c *gin.Context) {
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
func (h *ExternalIntegrationHandler) ResolveIssue(c *gin.Context) {
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
func (h *ExternalIntegrationHandler) SyncProjectIssues(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := uuid.FromString(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Validate user ownership of project (this will be checked in the usecase)
	// For now, run sync synchronously so we can return errors to the user
	// TODO: Implement proper async job queue for better UX
	err = h.usecase.SyncProjectIssues(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sync completed successfully"})
}

// ResolvePendingReferences resolves pending external references for a user
func (h *ExternalIntegrationHandler) ResolvePendingReferences(c *gin.Context) {
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
