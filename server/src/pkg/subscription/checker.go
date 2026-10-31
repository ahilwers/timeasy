// Package subscription provides subscription checking against RevenueCat. It is
// intentionally self-contained so that self-hosted instances can run without
// any billing configuration.
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
	// userID is the Keycloak user ID, which is also the RevenueCat App User ID
	// on every platform (web, iOS, Android). Returns (false, nil) only when the
	// provider definitively reports no active subscription. Transient provider
	// failures are returned as (false, err) so callers can fail open instead of
	// locking users out.
	CheckSubscription(ctx context.Context, userID string) (bool, error)

	// RefreshSubscription re-checks the subscription state bypassing any cache.
	RefreshSubscription(ctx context.Context, userID string) (bool, error)
}
