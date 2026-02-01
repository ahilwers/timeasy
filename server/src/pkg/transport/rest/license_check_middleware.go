package rest

import (
	"log/slog"
	"net/http"
	"timeasy-server/pkg/license"

	"github.com/gin-gonic/gin"
)

type licenseCheckMiddleware struct {
	licenseClient *license.LicenseManagerClient
}

// LicenseMiddleware defines the interface for license checking middleware
type LicenseMiddleware interface {
	HandlerFunc() gin.HandlerFunc
}

// NewLicenseCheckMiddleware creates a new license check middleware
func NewLicenseCheckMiddleware(licenseClient *license.LicenseManagerClient) LicenseMiddleware {
	return &licenseCheckMiddleware{
		licenseClient: licenseClient,
	}
}

func (mw *licenseCheckMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip license check if license manager is not configured
		if !mw.licenseClient.IsConfigured() {
			c.Next()
			return
		}

		// Get the auth token from context (set by JWT middleware)
		authToken, ok := GetAuthTokenFromContext(c)
		if !ok {
			slog.Error("Auth token not found in context for license check",
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		// Get user ID from the auth token
		userID, err := authToken.GetUserId()
		if err != nil {
			slog.Error("Failed to get user ID from token",
				"error", err,
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			c.String(http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		// Check subscription status
		active, err := mw.licenseClient.CheckSubscription(userID.String())
		if err != nil {
			slog.Error("Failed to check subscription",
				"error", err,
				"userID", userID.String(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			// On error, allow access but log the issue
			// This prevents license manager outages from blocking all users
			c.Next()
			return
		}

		if !active {
			slog.Warn("User does not have an active subscription",
				"userID", userID.String(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":   "subscription_inactive",
				"message": "Your subscription is not active. Please check your license.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
