package subscription

import "context"

// NoopChecker is used on self-hosted instances without billing configuration.
// It never blocks requests and reports itself as disabled so that clients hide
// all billing UI.
type NoopChecker struct{}

func NewNoopChecker() *NoopChecker {
	return &NoopChecker{}
}

func (c *NoopChecker) IsEnabled() bool {
	return false
}

func (c *NoopChecker) CheckSubscription(ctx context.Context, userID string) (bool, error) {
	return true, nil
}

func (c *NoopChecker) RefreshSubscription(ctx context.Context, userID string) (bool, error) {
	return true, nil
}
