// Package subscription provides subscription checking against external billing
// providers (Polar). It is intentionally self-contained so that self-hosted
// instances can run without any billing configuration.
package subscription

import "context"

// Checker reports whether a user has an active subscription.
type Checker interface {
	// IsEnabled reports whether subscription checking is active on this
	// instance. When false (self-hosted mode without billing configuration),
	// CheckSubscription always allows access and the billing endpoints report
	// themselves as disabled so that clients hide all billing UI.
	IsEnabled() bool

	// CheckSubscription reports whether the user has an active subscription.
	// userID is the Keycloak user ID; email is the user's Keycloak email and
	// is used as a fallback to match Polar customers created outside the app
	// funnel. Returns (false, nil) only when the provider definitively reports
	// no active subscription. Transient provider failures are returned as
	// (false, err) so callers can fail open instead of locking users out.
	CheckSubscription(ctx context.Context, userID, email string) (bool, error)

	// RefreshSubscription re-checks the subscription state bypassing any cache.
	RefreshSubscription(ctx context.Context, userID, email string) (bool, error)

	// GetPolarCustomerID resolves the internal Polar customer ID for the user.
	// Returns "" when no Polar customer exists for the user.
	GetPolarCustomerID(ctx context.Context, userID, email string) (string, error)
}

// BillingActions is implemented by checkers that can create Polar hosted
// flows (checkout, customer portal). Handlers type-assert to this interface
// and degrade gracefully when the active checker doesn't support it.
type BillingActions interface {
	// CreateCheckoutSession creates a hosted checkout session for the user and
	// returns its URL. The session links the Polar customer to userID via
	// external_customer_id. locale is an optional BCP 47 language tag (e.g.
	// "de") used for the checkout UI; empty means Polar's default (English).
	CreateCheckoutSession(ctx context.Context, userID, email, successURL, locale string) (string, error)

	// CreatePortalSession creates a short-lived Polar customer portal session
	// for the user and returns the portal URL, or "" when the user has no
	// Polar customer yet. The URL is short-lived and must be generated fresh
	// on every request.
	//
	// Note: Polar's customer portal is English-only (no locale parameter, and
	// the hosted page is rendered with lang="en"), so no locale is passed.
	CreatePortalSession(ctx context.Context, userID, email string) (string, error)
}
