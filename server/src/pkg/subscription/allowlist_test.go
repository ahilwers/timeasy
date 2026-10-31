package subscription

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeInnerChecker records calls so tests can assert delegation behavior.
type fakeInnerChecker struct {
	checkCalls   int
	refreshCalls int
	checkResult  bool
	checkErr     error
}

func (f *fakeInnerChecker) IsEnabled() bool { return true }

func (f *fakeInnerChecker) CheckSubscription(ctx context.Context, userID string) (bool, error) {
	f.checkCalls++
	return f.checkResult, f.checkErr
}

func (f *fakeInnerChecker) RefreshSubscription(ctx context.Context, userID string) (bool, error) {
	f.refreshCalls++
	return f.checkResult, f.checkErr
}

const allowlistedUserID = "aaaa1111-2222-3333-4444-555566667777"

func TestAllowlistChecker_AllowlistedUserGetsFullAccessWithoutBillingLookup(t *testing.T) {
	inner := &fakeInnerChecker{checkResult: false}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})

	active, err := checker.CheckSubscription(context.Background(), allowlistedUserID)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Zero(t, inner.checkCalls, "allowlisted users must not hit the wrapped checker")

	// Refresh must behave the same way.
	active, err = checker.RefreshSubscription(context.Background(), allowlistedUserID)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Zero(t, inner.refreshCalls)
}

func TestAllowlistChecker_OtherUsersAreDelegated(t *testing.T) {
	inner := &fakeInnerChecker{checkResult: false}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)
	assert.Equal(t, 1, inner.checkCalls)

	active, err = checker.RefreshSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)
	assert.Equal(t, 1, inner.refreshCalls)
}

func TestAllowlistChecker_TrimsWhitespaceAndSkipsEmptyEntries(t *testing.T) {
	inner := &fakeInnerChecker{checkResult: true}
	checker := NewAllowlistChecker(inner, []string{" " + allowlistedUserID + " ", "", "  "})

	active, err := checker.CheckSubscription(context.Background(), allowlistedUserID)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Zero(t, inner.checkCalls)
}

func TestAllowlistChecker_DelegatesIsEnabled(t *testing.T) {
	inner := &fakeInnerChecker{}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})
	assert.True(t, checker.IsEnabled())
}
