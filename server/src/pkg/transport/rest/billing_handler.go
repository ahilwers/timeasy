package rest

import (
	"log/slog"
	"net/http"
	"strings"
	"timeasy-server/pkg/subscription"

	"github.com/gin-gonic/gin"
)

// BillingHandler exposes the hosted billing flows: upgrade (Polar checkout),
// subscription status, and the Polar customer portal. All endpoints live
// outside the subscription-check middleware: a user without an active
// subscription must still be able to reach them, otherwise they could never
// start paying.
type BillingHandler struct {
	checker      subscription.Checker
	webclientURL string // base URL of the web client, used for the checkout success redirect
}

func NewBillingHandler(checker subscription.Checker, webclientURL string) *BillingHandler {
	return &BillingHandler{
		checker:      checker,
		webclientURL: strings.TrimSuffix(webclientURL, "/"),
	}
}

// GetBillingStatus returns whether billing (subscriptions) is enabled on this
// instance. Public endpoint so clients can hide billing UI on self-hosted
// instances without authenticating first.
// GET /api/v1/billing/status
func (h *BillingHandler) GetBillingStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"enabled": h.checker.IsEnabled()})
}

// StartUpgrade creates a Polar hosted checkout session for the current user
// and returns its URL. The session links the Polar customer to the Keycloak
// user ID via external_customer_id.
// POST /api/v1/upgrade  (optional body: {"locale": "de"})
func (h *BillingHandler) StartUpgrade(c *gin.Context) {
	if !h.checker.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{"enabled": false, "url": nil})
		return
	}
	actions, ok := h.checker.(subscription.BillingActions)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"enabled": true, "url": nil})
		return
	}
	userID, email, ok := h.userFromContext(c)
	if !ok {
		return // response already written
	}
	var successURL string
	if h.webclientURL != "" {
		successURL = h.webclientURL + "/upgrade/success"
	}
	checkoutURL, err := actions.CreateCheckoutSession(c.Request.Context(), userID, email, successURL, h.checkoutLocale(c))
	if err != nil {
		slog.Error("Failed to create polar checkout session", "error", err, "userID", userID)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "billing_unavailable",
			"message": "Could not start the upgrade flow. Please try again later.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": true, "url": checkoutURL})
}

// GetSubscriptionStatus force-refreshes the current user's subscription state
// (bypassing the cache), so clients can confirm a just-completed payment.
// GET /api/v1/subscription/status
func (h *BillingHandler) GetSubscriptionStatus(c *gin.Context) {
	userID, email, ok := h.userFromContext(c)
	if !ok {
		return // response already written
	}
	active, err := h.checker.RefreshSubscription(c.Request.Context(), userID, email)
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

// CreatePortalSession returns a short-lived URL to the Polar customer portal
// where the user can manage payment methods, invoices, and cancellation.
// The URL is generated fresh on every request (customer sessions are
// short-lived). When the user has no Polar customer yet, url is null and the
// client should offer the upgrade flow instead.
// POST /api/v1/billing/portal
// Note: Polar's hosted customer portal is English-only; no locale is passed.
func (h *BillingHandler) CreatePortalSession(c *gin.Context) {
	if !h.checker.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{"enabled": false, "url": nil})
		return
	}
	actions, ok := h.checker.(subscription.BillingActions)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"enabled": true, "url": nil})
		return
	}
	userID, email, ok := h.userFromContext(c)
	if !ok {
		return // response already written
	}
	portalURL, err := actions.CreatePortalSession(c.Request.Context(), userID, email)
	if err != nil {
		slog.Error("Failed to create polar customer portal session", "error", err, "userID", userID)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "billing_unavailable",
			"message": "Could not open the subscription portal. Please try again later.",
		})
		return
	}
	var url any
	if portalURL != "" {
		url = portalURL
	}
	c.JSON(http.StatusOK, gin.H{"enabled": true, "url": url})
}

// checkoutLocale determines the language for the hosted checkout: the client
// sends its active UI language in the request body, with the browser's
// Accept-Language header as fallback. Empty means Polar's default (English).
func (h *BillingHandler) checkoutLocale(c *gin.Context) string {
	var req struct {
		Locale string `json:"locale"`
	}
	// The body is optional; ignore binding errors (e.g. empty body).
	_ = c.ShouldBindJSON(&req)
	if req.Locale != "" {
		return req.Locale
	}
	return localeFromAcceptLanguage(c.GetHeader("Accept-Language"))
}

// localeFromAcceptLanguage extracts the preferred language tag from an
// Accept-Language header, e.g. "de-DE,de;q=0.9,en;q=0.8" → "de-DE".
func localeFromAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if tag != "" && tag != "*" {
			return tag
		}
	}
	return ""
}

func (h *BillingHandler) userFromContext(c *gin.Context) (string, string, bool) {
	authToken, ok := GetAuthTokenFromContext(c)
	if !ok {
		slog.Error("Auth token not found in context for billing request",
			"path", c.Request.URL.Path,
			"method", c.Request.Method)
		c.String(http.StatusUnauthorized, "Unauthorized")
		return "", "", false
	}
	userID, err := authToken.GetUserId()
	if err != nil {
		slog.Error("Failed to get user ID from token", "error", err)
		c.String(http.StatusUnauthorized, "Unauthorized")
		return "", "", false
	}
	email, emailErr := authToken.GetEmail()
	if emailErr != nil {
		slog.Debug("Auth token without email claim", "userID", userID)
	}
	return userID.String(), email, true
}
