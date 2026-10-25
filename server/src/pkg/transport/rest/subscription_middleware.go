package rest

import (
	"log/slog"
	"net/http"
	"timeasy-server/pkg/subscription"

	"github.com/gin-gonic/gin"
)

// SubscriptionMiddleware defines the interface for subscription checking middleware.
type SubscriptionMiddleware interface {
	HandlerFunc() gin.HandlerFunc
}

type subscriptionCheckMiddleware struct {
	checker subscription.Checker
}

// NewSubscriptionCheckMiddleware creates a new subscription check middleware.
func NewSubscriptionCheckMiddleware(checker subscription.Checker) SubscriptionMiddleware {
	return &subscriptionCheckMiddleware{
		checker: checker,
	}
}

func (mw *subscriptionCheckMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Self-hosted instances without billing configuration allow everything.
		if !mw.checker.IsEnabled() {
			c.Next()
			return
		}

		// Get the auth token from context (set by JWT middleware)
		authToken, ok := GetAuthTokenFromContext(c)
		if !ok {
			slog.Error("Auth token not found in context for subscription check",
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

		// Email is optional (used as fallback for Polar customer matching)
		email, _ := authToken.GetEmail()

		// Check subscription status
		active, err := mw.checker.CheckSubscription(c.Request.Context(), userID.String(), email)
		if err != nil {
			slog.Error("Failed to check subscription",
				"error", err,
				"userID", userID.String(),
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
			// On transient errors, allow access but log the issue
			// This prevents billing provider outages from blocking all users
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
				"message": "Your subscription is not active. Please subscribe to timeasy Pro.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
