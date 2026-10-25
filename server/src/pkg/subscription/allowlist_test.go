package subscription

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeInnerChecker records calls so tests can assert delegation behavior.
type fakeInnerChecker struct {
	checkCalls          int
	refreshCalls        int
	customerLookupCalls int
	checkResult         bool
	checkErr            error
	customerID          string
}

func (f *fakeInnerChecker) IsEnabled() bool { return true }

func (f *fakeInnerChecker) CheckSubscription(ctx context.Context, userID, email string) (bool, error) {
	f.checkCalls++
	return f.checkResult, f.checkErr
}

func (f *fakeInnerChecker) RefreshSubscription(ctx context.Context, userID, email string) (bool, error) {
	f.refreshCalls++
	return f.checkResult, f.checkErr
}

func (f *fakeInnerChecker) GetPolarCustomerID(ctx context.Context, userID, email string) (string, error) {
	f.customerLookupCalls++
	return f.customerID, nil
}

const allowlistedUserID = "aaaa1111-2222-3333-4444-555566667777"

func TestAllowlistChecker_AllowlistedUserGetsFullAccessWithoutPolar(t *testing.T) {
	inner := &fakeInnerChecker{checkResult: false}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})

	active, err := checker.CheckSubscription(context.Background(), allowlistedUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Zero(t, inner.checkCalls, "allowlisted users must not hit the wrapped checker")

	// Refresh must behave the same way.
	active, err = checker.RefreshSubscription(context.Background(), allowlistedUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Zero(t, inner.refreshCalls)
}

func TestAllowlistChecker_OtherUsersAreDelegated(t *testing.T) {
	inner := &fakeInnerChecker{checkResult: false, customerID: testCustomerID}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.False(t, active)
	assert.Equal(t, 1, inner.checkCalls)

	// No Polar customer for allowlisted users, delegation for everyone else.
	customerID, err := checker.GetPolarCustomerID(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Equal(t, testCustomerID, customerID)
	assert.Equal(t, 1, inner.customerLookupCalls)
}

func TestAllowlistChecker_AllowlistedUsersHaveNoPolarCustomer(t *testing.T) {
	inner := &fakeInnerChecker{customerID: testCustomerID}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})

	customerID, err := checker.GetPolarCustomerID(context.Background(), allowlistedUserID, testEmail)
	require.NoError(t, err)
	assert.Empty(t, customerID)
	assert.Zero(t, inner.customerLookupCalls)
}

func TestAllowlistChecker_TrimsWhitespaceAndSkipsEmptyEntries(t *testing.T) {
	inner := &fakeInnerChecker{checkResult: true}
	checker := NewAllowlistChecker(inner, []string{" " + allowlistedUserID + " ", "", "  "})

	active, err := checker.CheckSubscription(context.Background(), allowlistedUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Zero(t, inner.checkCalls)
}

func TestAllowlistChecker_DelegatesIsEnabled(t *testing.T) {
	inner := &fakeInnerChecker{}
	checker := NewAllowlistChecker(inner, []string{allowlistedUserID})
	assert.True(t, checker.IsEnabled())
}
