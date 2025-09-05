package rest

import (
	"net/http"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

type UserExternalAccountHandler interface {
	CreateAccount(context *gin.Context)
	GetAccounts(context *gin.Context)
	GetAccountsByProvider(context *gin.Context)
	UpdateAccount(context *gin.Context)
	DeleteAccount(context *gin.Context)
	TestAccount(context *gin.Context)
}

type userExternalAccountHandler struct {
	tokenVerifier TokenVerifier
	usecase       *usecase.UserExternalAccountUseCase
}

func NewUserExternalAccountHandler(tokenVerifier TokenVerifier, usecase *usecase.UserExternalAccountUseCase) UserExternalAccountHandler {
	return &userExternalAccountHandler{
		tokenVerifier: tokenVerifier,
		usecase:       usecase,
	}
}

// CreateAccount creates a new external account for the user
func (h *userExternalAccountHandler) CreateAccount(c *gin.Context) {
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

	var req model.UserExternalAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.usecase.CreateAccount(c.Request.Context(), userUUID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, account)
}

// UpdateAccount updates an existing external account
func (h *userExternalAccountHandler) UpdateAccount(c *gin.Context) {
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

	accountIDStr := c.Param("accountId")
	accountID, err := uuid.FromString(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	var req model.UserExternalAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.usecase.UpdateAccount(c.Request.Context(), userUUID, accountID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, account)
}

// DeleteAccount deletes an external account
func (h *userExternalAccountHandler) DeleteAccount(c *gin.Context) {
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

	accountIDStr := c.Param("accountId")
	accountID, err := uuid.FromString(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	if err := h.usecase.DeleteAccount(c.Request.Context(), userUUID, accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

// GetAccounts returns all external accounts for the user
func (h *userExternalAccountHandler) GetAccounts(c *gin.Context) {
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

	accounts, err := h.usecase.GetAccountsByUser(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// GetAccountsByProvider returns external accounts for a specific provider
func (h *userExternalAccountHandler) GetAccountsByProvider(c *gin.Context) {
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

	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}

	if !model.IsValidProvider(provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	accounts, err := h.usecase.GetAccountsByUserAndProvider(c.Request.Context(), userUUID, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// TestAccount tests the connection to an external account
func (h *userExternalAccountHandler) TestAccount(c *gin.Context) {
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

	accountIDStr := c.Param("accountId")
	accountID, err := uuid.FromString(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID"})
		return
	}

	if err := h.usecase.TestAccount(c.Request.Context(), userUUID, accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account connection successful"})
}