package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID        = "11111111-1111-1111-1111-111111111111"
	testAPIKey        = "test-secret-key"
	testProjectID     = "proj_test123"
	testEntitlementID = "entl_test123"
)

// fakeRevenueCat is a configurable RevenueCat API v2 fake for tests.
type fakeRevenueCat struct {
	server       *httptest.Server
	requestCount atomic.Int32
	handler      func(w http.ResponseWriter, r *http.Request)
}

func newFakeRevenueCat(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *fakeRevenueCat {
	f := &fakeRevenueCat{handler: handler}
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/projects/", func(w http.ResponseWriter, r *http.Request) {
		f.requestCount.Add(1)
		f.handler(w, r)
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeRevenueCat) baseURL() string {
	return f.server.URL + "/v2"
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// fakeClock is an injectable clock for TTL and expiry tests.
type fakeClock struct {
	current atomic.Int64
}

func newFakeClock() *fakeClock {
	clock := &fakeClock{}
	clock.current.Store(time.Now().Unix())
	return clock
}

func (c *fakeClock) now() time.Time {
	return time.Unix(c.current.Load(), 0)
}

func (c *fakeClock) advance(d time.Duration) {
	c.current.Add(int64(d.Seconds()))
}

func newTestChecker(baseURL string, clock *fakeClock) *RevenueCatChecker {
	checker := NewRevenueCatChecker(RevenueCatConfig{
		APIKey:        testAPIKey,
		ProjectID:     testProjectID,
		EntitlementID: testEntitlementID,
		CacheTTL:      15 * time.Minute,
		BaseURL:       baseURL,
	})
	checker.now = clock.now
	return checker
}

func int64Ptr(value int64) *int64 {
	return &value
}

// customerResponse builds a RevenueCat v2 customer response. A nil expiry
// represents an entitlement without a fixed expiry.
func customerResponse(entitlementID string, expiresAt *int64) map[string]any {
	items := []map[string]any{}
	if entitlementID != "" {
		items = append(items, map[string]any{
			"object":         "customer.active_entitlement",
			"entitlement_id": entitlementID,
			"expires_at":     expiresAt,
		})
	}
	return map[string]any{
		"object":     "customer",
		"id":         testUserID,
		"project_id": testProjectID,
		"active_entitlements": map[string]any{
			"object": "list",
			"items":  items,
		},
	}
}

func TestRevenueCatChecker_UsesV2CustomerEndpointAndActiveEntitlement(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).UnixMilli()
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, fmt.Sprintf("/v2/projects/%s/customers/%s", testProjectID, testUserID), r.URL.Path)
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, int64Ptr(future)))
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)

	// Second call must be served from cache without hitting the API again.
	requestsBefore := fake.requestCount.Load()
	active, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())
}

func TestRevenueCatChecker_ExpiredEntitlementIsInactive(t *testing.T) {
	past := time.Now().Add(-time.Hour).UnixMilli()
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, int64Ptr(past)))
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)
}

func TestRevenueCatChecker_LifetimeEntitlementIsActive(t *testing.T) {
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, nil))
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)
}

func TestRevenueCatChecker_OtherEntitlementDoesNotGrantAccess(t *testing.T) {
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, customerResponse("entl_some_other", nil))
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)
}

func TestRevenueCatChecker_NoActiveEntitlementsIsInactive(t *testing.T) {
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, customerResponse("", nil))
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)
}

func TestRevenueCatChecker_UnknownUserIsInactiveAndCached(t *testing.T) {
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"type": "resource_missing"})
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)

	// Definitive negative results are cached too.
	requestsBefore := fake.requestCount.Load()
	active, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())
}

// A negative result must expire quickly: right after a purchase the app must
// not be locked out by a stale "no subscription" answer from the cache.
func TestRevenueCatChecker_NegativeResultExpiresQuickly(t *testing.T) {
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"type": "resource_missing"})
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.False(t, active)

	// Still cached directly after the check.
	requestsBefore := fake.requestCount.Load()
	_, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())

	// After the negative TTL the check asks RevenueCat again.
	clock.advance(DefaultNegativeCacheTTL + time.Second)
	_, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Greater(t, fake.requestCount.Load(), requestsBefore)
}

func TestRevenueCatChecker_TransientErrorIsNotCached(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	future := time.Now().Add(24 * time.Hour).UnixMilli()
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"type": "server_error"})
			return
		}
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, int64Ptr(future)))
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	assert.Error(t, err)
	assert.False(t, active)

	// Transient failures must not be cached; the next call retries.
	fail.Store(false)
	active, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)
}

func TestRevenueCatChecker_SendsBearerKeyAndScopedProjectPath(t *testing.T) {
	var receivedAuth string
	var receivedPath string
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedPath = r.URL.Path
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, nil))
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	_, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, "Bearer "+testAPIKey, receivedAuth)
	assert.Equal(t, fmt.Sprintf("/v2/projects/%s/customers/%s", testProjectID, testUserID), receivedPath)
}

func TestRevenueCatChecker_CacheExpiryRechecks(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).UnixMilli()
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, int64Ptr(future)))
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	_, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	requestsBefore := fake.requestCount.Load()

	clock.advance(14 * time.Minute)
	_, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())

	clock.advance(2 * time.Minute)
	_, err = checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Greater(t, fake.requestCount.Load(), requestsBefore)
}

func TestRevenueCatChecker_RefreshSubscriptionBypassesCache(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).UnixMilli()
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, int64Ptr(future)))
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	_, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	requestsBefore := fake.requestCount.Load()

	active, err := checker.RefreshSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Greater(t, fake.requestCount.Load(), requestsBefore)
}

func TestRevenueCatChecker_ConcurrentChecksCollapseToSingleFlight(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).UnixMilli()
	fake := newFakeRevenueCat(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		writeJSON(w, http.StatusOK, customerResponse(testEntitlementID, int64Ptr(future)))
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	const concurrent = 10
	var wg sync.WaitGroup
	results := make([]bool, concurrent)
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			active, err := checker.CheckSubscription(context.Background(), testUserID)
			require.NoError(t, err)
			results[idx] = active
		}(i)
	}
	wg.Wait()

	for i, active := range results {
		assert.True(t, active, "goroutine %d should see an active subscription", i)
	}
	assert.LessOrEqual(t, fake.requestCount.Load(), int32(1),
		"concurrent checks must collapse into a single RevenueCat API call")
}

func TestRevenueCatConfig_RequiredValuesArePassedThrough(t *testing.T) {
	checker := NewRevenueCatChecker(RevenueCatConfig{
		APIKey:        testAPIKey,
		ProjectID:     testProjectID,
		EntitlementID: testEntitlementID,
	})
	assert.Equal(t, testProjectID, checker.projectID)
	assert.Equal(t, testEntitlementID, checker.entitlementID)
	assert.Equal(t, DefaultCacheTTL, checker.cacheTTL)
	assert.Equal(t, revenueCatBaseURL, checker.baseURL)
	assert.True(t, checker.IsEnabled())
}

func TestNoopChecker(t *testing.T) {
	checker := NewNoopChecker()
	assert.False(t, checker.IsEnabled())

	active, err := checker.CheckSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)

	active, err = checker.RefreshSubscription(context.Background(), testUserID)
	require.NoError(t, err)
	assert.True(t, active)
}
