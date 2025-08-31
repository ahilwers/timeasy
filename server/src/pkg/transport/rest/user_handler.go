package rest

import (
	"fmt"
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/usecase"
)

type UserHandler struct {
	userUsecase   usecase.UserUsecase
	tokenVerifier TokenVerifier
}

func NewUserHandler(userUsecase usecase.UserUsecase, tokenVerifier TokenVerifier) *UserHandler {
	return &UserHandler{
		userUsecase:   userUsecase,
		tokenVerifier: tokenVerifier,
	}
}

func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	userRoutes := router.Group("/user")
	{
		userRoutes.GET("/profile", h.GetUserProfile)
		userRoutes.PUT("/profile", h.UpdateUserProfile)
		userRoutes.POST("/change-password", h.ChangePassword)
	}
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// Verify token but also extract the raw token string for Keycloak Account API
	_, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		glog.Errorf("error verifying token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	rawToken, err := h.extractRawToken(c)
	if err != nil {
		glog.Errorf("error extracting raw token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token format"})
		return
	}
	
	user, err := h.userUsecase.GetUserProfile(c.Request.Context(), rawToken)
	if err != nil {
		glog.Errorf("error getting user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}
	
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	// Verify token but also extract the raw token string for Keycloak Account API
	_, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		glog.Errorf("error verifying token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	rawToken, err := h.extractRawToken(c)
	if err != nil {
		glog.Errorf("error extracting raw token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token format"})
		return
	}
	
	var updateRequest model.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		glog.Errorf("error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	user, err := h.userUsecase.UpdateUserProfile(c.Request.Context(), rawToken, &updateRequest)
	if err != nil {
		glog.Errorf("error updating user profile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}
	
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	// Verify token but also extract the raw token string for Keycloak Account API
	_, err := h.tokenVerifier.VerifyToken(c)
	if err != nil {
		glog.Errorf("error verifying token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	rawToken, err := h.extractRawToken(c)
	if err != nil {
		glog.Errorf("error extracting raw token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token format"})
		return
	}
	
	var passwordRequest model.PasswordChangeRequest
	if err := c.ShouldBindJSON(&passwordRequest); err != nil {
		glog.Errorf("error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	err = h.userUsecase.ChangePassword(c.Request.Context(), rawToken, &passwordRequest)
	if err != nil {
		glog.Errorf("error changing password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// extractRawToken extracts the Bearer token from the Authorization header
func (h *UserHandler) extractRawToken(c *gin.Context) (string, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}
	
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("malformed authorization header")
	}
	
	return parts[1], nil
}