package rest

import (
	"log/slog"
	"net/http"
	"strings"
	"timeasy-server/pkg/subscription"

	"github.com/gin-gonic/gin"
)

// BillingHandler exposes subscription status endpoints. They live outside the
// subscription-check middleware: a user without an active subscription must
// still be able to check their status, and clients use the public status
// endpoint to decide whether to show any billing UI at all.
//
// Checkout and subscription management do not need server endpoints: the
// clients use the RevenueCat SDKs directly (web: Paddle checkout plus
// CustomerInfo.managementURL, mobile: store purchases).
type BillingHandler struct {
	checker   subscription.Checker
	webAPIKey string // public RevenueCat Web Billing SDK key, sent to web clients
}

func NewBillingHandler(checker subscription.Checker, webAPIKey string) *BillingHandler {
	return &BillingHandler{
		checker:   checker,
		webAPIKey: strings.TrimSpace(webAPIKey),
	}
}

// GetBillingStatus returns whether billing (subscriptions) is enabled on this
// instance. Public endpoint so clients can hide billing UI on self-hosted
// instances without authenticating first. When billing is enabled, the
// instance's public RevenueCat Web Billing key is included so the web client
// does not need a key baked into its build.
// GET /api/v1/billing/status
func (h *BillingHandler) GetBillingStatus(c *gin.Context) {
	response := gin.H{"enabled": h.checker.IsEnabled()}
	if h.checker.IsEnabled() && h.webAPIKey != "" {
		response["webApiKey"] = h.webAPIKey
	}
	c.JSON(http.StatusOK, response)
}

// GetSubscriptionStatus force-refreshes the current user's subscription state
// (bypassing the cache), so clients can confirm a just-completed payment.
// GET /api/v1/subscription/status
func (h *BillingHandler) GetSubscriptionStatus(c *gin.Context) {
	userID, ok := h.userFromContext(c)
	if !ok {
		return // response already written
	}
	active, err := h.checker.RefreshSubscription(c.Request.Context(), userID)
	if err != nil {
		slog.Error("Failed to refresh subscription status", "error", err, "userID", userID)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "billing_unavailable",
			"message": "Could not determine the subscription status. Please try again later.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": h.checker.IsEnabled(), "active": active})
}

func (h *BillingHandler) userFromContext(c *gin.Context) (string, bool) {
	authToken, ok := GetAuthTokenFromContext(c)
	if !ok {
		slog.Error("Auth token not found in context for billing request",
			"path", c.Request.URL.Path,
			"method", c.Request.Method)
		c.String(http.StatusUnauthorized, "Unauthorized")
		return "", false
	}
	userID, err := authToken.GetUserId()
	if err != nil {
		slog.Error("Failed to get user ID from token", "error", err)
		c.String(http.StatusUnauthorized, "Unauthorized")
		return "", false
	}
	return userID.String(), true
}
