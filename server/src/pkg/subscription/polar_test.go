package subscription

import (
	"context"
	"encoding/json"
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
	testUserID      = "11111111-1111-1111-1111-111111111111"
	testEmail       = "user@example.com"
	testCustomerID  = "992fae2a-2a17-4b7a-8d9e-e287cf90131b"
	testProductID   = "d8dd2de1-21b7-4a41-8bc3-ce909c0cfe23"
	testCheckoutURL = "https://sandbox.polar.sh/checkout"
	testPortalURL   = "https://sandbox.polar.sh/portal/session"
)

// fakePolar is a configurable Polar API fake for tests.
type fakePolar struct {
	server       *httptest.Server
	requestCount atomic.Int32

	externalCustomer func(w http.ResponseWriter, r *http.Request)
	emailList        func(w http.ResponseWriter, r *http.Request)
	customerState    func(w http.ResponseWriter, r *http.Request)
	checkout         func(w http.ResponseWriter, r *http.Request)
	portalSession    func(w http.ResponseWriter, r *http.Request)

	backfillCalled atomic.Bool
	backfillBody   atomic.Value
}

func newFakePolar(t *testing.T, configure func(f *fakePolar)) *fakePolar {
	f := &fakePolar{}
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/customers/external/", func(w http.ResponseWriter, r *http.Request) {
		f.requestCount.Add(1)
		if f.externalCustomer != nil {
			f.externalCustomer(w, r)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "ResourceNotFound", "detail": ""})
	})

	mux.HandleFunc("/v1/customers", func(w http.ResponseWriter, r *http.Request) {
		f.requestCount.Add(1)
		if f.emailList != nil {
			f.emailList(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": []map[string]any{}})
	})

	mux.HandleFunc("/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/customers/"+testCustomerID+"/state" && r.Method == http.MethodGet:
			f.requestCount.Add(1)
			if f.customerState != nil {
				f.customerState(w, r)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"id":                   testCustomerID,
				"active_subscriptions": []map[string]any{},
			})
		case r.URL.Path == "/v1/customers/"+testCustomerID && r.Method == http.MethodPatch:
			f.requestCount.Add(1)
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			f.backfillBody.Store(body)
			f.backfillCalled.Store(true)
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "ResourceNotFound", "detail": ""})
		}
	})

	mux.HandleFunc("/v1/checkouts/", func(w http.ResponseWriter, r *http.Request) {
		f.requestCount.Add(1)
		if f.checkout != nil {
			f.checkout(w, r)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"id": "checkout_1", "url": testCheckoutURL})
	})

	mux.HandleFunc("/v1/customer-sessions/", func(w http.ResponseWriter, r *http.Request) {
		f.requestCount.Add(1)
		if f.portalSession != nil {
			f.portalSession(w, r)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"id": "session_1", "customer_portal_url": testPortalURL})
	})

	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	if configure != nil {
		configure(f)
	}
	return f
}

func (f *fakePolar) baseURL() string {
	return f.server.URL
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func newState(customerID, productID, status string) map[string]any {
	return map[string]any{
		"id": customerID,
		"active_subscriptions": []map[string]any{
			{"id": "e5149aae-e521-42b9-b24c-abb3d71eea2e", "status": status, "product_id": productID},
		},
	}
}

// fakeClock is an injectable clock for TTL tests.
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

func newTestChecker(baseURL string, clock *fakeClock) *PolarChecker {
	checker := NewPolarChecker(PolarConfig{
		Env:         PolarEnvSandbox,
		ProductID:   testProductID,
		AccessToken: "test-token",
		CacheTTL:    15 * time.Minute,
	})
	checker.baseURL = baseURL
	checker.now = clock.now
	return checker
}

func TestPolarChecker_CheckSubscription_ActiveCustomer(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "active"))
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)

	// Second call must be served from cache without hitting the API again.
	requestsBefore := fake.requestCount.Load()
	active, err = checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())
}

func TestPolarChecker_CheckSubscription_NoCustomer(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		// External lookup 404 and empty email list → definitively no customer.
		f.emailList = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"items": []map[string]any{}})
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.False(t, active)

	// Negative results are cached too (no extra API calls within TTL).
	requestsBefore := fake.requestCount.Load()
	active, err = checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.False(t, active)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())
}

func TestPolarChecker_CheckSubscription_EmailFallbackBackfillsExternalID(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		// No customer with the external ID (created outside the app funnel)...
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "ResourceNotFound", "detail": ""})
		}
		// ...but one with the same email.
		f.emailList = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{
				"items": []map[string]any{
					{"id": testCustomerID, "email": testEmail},
				},
			})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "active"))
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
	assert.True(t, fake.backfillCalled.Load(), "external_id should be backfilled")
	backfill := fake.backfillBody.Load().(map[string]any)
	assert.Equal(t, testUserID, backfill["external_id"])

	// The resolved customer ID must now be available for portal sessions.
	customerID, err := checker.GetPolarCustomerID(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Equal(t, testCustomerID, customerID)
}

func TestPolarChecker_CheckSubscription_EmailMatchRequiresExactEmail(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.emailList = func(w http.ResponseWriter, r *http.Request) {
			// Polar might return candidates when the email filter is ignored;
			// only an exact match may be used.
			writeJSON(w, http.StatusOK, map[string]any{
				"items": []map[string]any{
					{"id": testCustomerID, "email": "someone.else@example.com"},
				},
			})
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.False(t, active)
}

func TestPolarChecker_CheckSubscription_IgnoresOtherProducts(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			// Active subscription for a different product must not count.
			writeJSON(w, http.StatusOK, newState(testCustomerID, "00000000-0000-0000-0000-000000000000", "active"))
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.False(t, active)
}

func TestPolarChecker_CheckSubscription_TrialingCountsAsActive(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "trialing"))
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
}

func TestPolarChecker_CheckSubscription_TransientErrorFailsOpen(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			// Polar outage: a transient 5xx must surface as an error, not as
			// "no subscription" (callers fail open on errors).
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "boom"})
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	assert.Error(t, err)
	assert.False(t, active)
}

func TestPolarChecker_CheckSubscription_CacheExpiryRechecks(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "active"))
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	_, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	requestsBefore := fake.requestCount.Load()

	// Within the TTL: cached.
	clock.advance(14 * time.Minute)
	_, err = checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Equal(t, requestsBefore, fake.requestCount.Load())

	// After the TTL: re-check.
	clock.advance(2 * time.Minute)
	_, err = checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Greater(t, fake.requestCount.Load(), requestsBefore)
}

func TestPolarChecker_RefreshSubscription_BypassesCache(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "active"))
		}
	})
	clock := newFakeClock()
	checker := newTestChecker(fake.baseURL(), clock)

	_, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	requestsBefore := fake.requestCount.Load()

	// Force refresh must hit the API even within the TTL.
	active, err := checker.RefreshSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)
	assert.Greater(t, fake.requestCount.Load(), requestsBefore)
}

func TestPolarChecker_CreateCheckoutSession(t *testing.T) {
	var receivedBody map[string]any
	fake := newFakePolar(t, func(f *fakePolar) {
		f.checkout = func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/v1/checkouts/", r.URL.Path)
			require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody))
			writeJSON(w, http.StatusCreated, map[string]any{"id": "checkout_1", "url": testCheckoutURL})
		}
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	url, err := checker.CreateCheckoutSession(context.Background(), testUserID, testEmail, "https://web/upgrade/success", "de")
	require.NoError(t, err)
	assert.Equal(t, testCheckoutURL, url)
	assert.Equal(t, testUserID, receivedBody["external_customer_id"])
	assert.Equal(t, testEmail, receivedBody["customer_email"])
	assert.Equal(t, "https://web/upgrade/success", receivedBody["success_url"])
	assert.Equal(t, "de", receivedBody["locale"])
}

func TestPolarChecker_CreateCheckoutSession_OmitsInvalidLocale(t *testing.T) {
	var receivedBody map[string]any
	fake := newFakePolar(t, func(f *fakePolar) {
		f.checkout = func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&receivedBody))
			writeJSON(w, http.StatusCreated, map[string]any{"id": "checkout_1", "url": testCheckoutURL})
		}
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	_, err := checker.CreateCheckoutSession(context.Background(), testUserID, testEmail, "", "not a locale!")
	require.NoError(t, err)
	_, hasLocale := receivedBody["locale"]
	assert.False(t, hasLocale, "invalid locale must be omitted so Polar falls back to English")
}

func TestNormalizeLocale(t *testing.T) {
	assert.Equal(t, "de", normalizeLocale("de"))
	assert.Equal(t, "de-DE", normalizeLocale("de-DE"))
	assert.Equal(t, "en-GB-oxendict", normalizeLocale("en-GB-oxendict"))
	assert.Equal(t, "", normalizeLocale(""))
	assert.Equal(t, "", normalizeLocale("de_DE"))
	assert.Equal(t, "", normalizeLocale("not a locale!"))
	assert.Equal(t, "", normalizeLocale("x"))
}

func TestPolarChecker_CreatePortalSession_NoCustomer(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.emailList = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"items": []map[string]any{}})
		}
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	url, err := checker.CreatePortalSession(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Empty(t, url)
}

func TestPolarChecker_CreatePortalSession_ReturnsPortalURL(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "active"))
		}
		f.portalSession = func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/v1/customer-sessions/", r.URL.Path)
			writeJSON(w, http.StatusCreated, map[string]any{
				"id":                  "session_1",
				"customer_portal_url": testPortalURL,
			})
		}
	})
	checker := newTestChecker(fake.baseURL(), newFakeClock())

	url, err := checker.CreatePortalSession(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Equal(t, testPortalURL, url)
}

func TestPolarChecker_ConcurrentChecksCollapseToSingleFlight(t *testing.T) {
	fake := newFakePolar(t, func(f *fakePolar) {
		f.externalCustomer = func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"id": testCustomerID})
		}
		f.customerState = func(w http.ResponseWriter, r *http.Request) {
			// Small delay so the concurrent requests actually overlap.
			time.Sleep(50 * time.Millisecond)
			writeJSON(w, http.StatusOK, newState(testCustomerID, testProductID, "active"))
		}
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
			active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
			require.NoError(t, err)
			results[idx] = active
		}(i)
	}
	wg.Wait()

	for i, active := range results {
		assert.True(t, active, "goroutine %d should see an active subscription", i)
	}
	// One resolution = external lookup (1) + customer state (1). Without
	// single-flight, 10 concurrent misses would cause up to 20 API calls.
	assert.LessOrEqual(t, fake.requestCount.Load(), int32(2),
		"concurrent checks must collapse into a single Polar round trip")
}

func TestNoopChecker(t *testing.T) {
	checker := NewNoopChecker()
	assert.False(t, checker.IsEnabled())

	active, err := checker.CheckSubscription(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.True(t, active)

	customerID, err := checker.GetPolarCustomerID(context.Background(), testUserID, testEmail)
	require.NoError(t, err)
	assert.Empty(t, customerID)
}

func TestPolarBaseURL(t *testing.T) {
	assert.Equal(t, "https://sandbox-api.polar.sh", polarBaseURL(PolarEnvSandbox))
	assert.Equal(t, "https://api.polar.sh", polarBaseURL(PolarEnvProduction))
	assert.Equal(t, "https://sandbox-api.polar.sh", polarBaseURL("unknown-falls-back-to-sandbox"))
}
