package subscription

import (
	"context"
	"strings"
)

// AllowlistChecker grants full access to a configurable set of Keycloak user
// IDs without any Polar lookup — useful for the owner and testers of a hosted
// instance who should not pay for their own app. All other users are delegated
// to the wrapped Checker.
type AllowlistChecker struct {
	allowlist map[string]struct{}
	inner     Checker
}

func NewAllowlistChecker(inner Checker, userIDs []string) *AllowlistChecker {
	allowlist := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			allowlist[trimmed] = struct{}{}
		}
	}
	return &AllowlistChecker{allowlist: allowlist, inner: inner}
}

func (c *AllowlistChecker) IsEnabled() bool {
	return c.inner.IsEnabled()
}

func (c *AllowlistChecker) CheckSubscription(ctx context.Context, userID, email string) (bool, error) {
	if c.isAllowlisted(userID) {
		return true, nil
	}
	return c.inner.CheckSubscription(ctx, userID, email)
}

func (c *AllowlistChecker) RefreshSubscription(ctx context.Context, userID, email string) (bool, error) {
	if c.isAllowlisted(userID) {
		return true, nil
	}
	return c.inner.RefreshSubscription(ctx, userID, email)
}

func (c *AllowlistChecker) GetPolarCustomerID(ctx context.Context, userID, email string) (string, error) {
	// Allowlisted users have no Polar customer; the billing endpoints treat
	// "" as "no customer" and offer the upgrade flow instead.
	if c.isAllowlisted(userID) {
		return "", nil
	}
	return c.inner.GetPolarCustomerID(ctx, userID, email)
}

func (c *AllowlistChecker) isAllowlisted(userID string) bool {
	_, ok := c.allowlist[userID]
	return ok
}
