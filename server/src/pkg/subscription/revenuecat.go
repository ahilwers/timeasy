package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	// revenueCatBaseURL is the base URL of the RevenueCat REST API v2.
	revenueCatBaseURL = "https://api.revenuecat.com/v2"
	// DefaultCacheTTL is the default duration a positive subscription check
	// result stays valid.
	DefaultCacheTTL = 15 * time.Minute
	// DefaultNegativeCacheTTL is how long a "no active subscription" answer is
	// cached. Deliberately short: after a purchase the app must not be rejected
	// with a stale 402 from the cache, so a fresh check happens quickly.
	DefaultNegativeCacheTTL = time.Minute
)

// RevenueCatConfig holds the configuration for the RevenueCatChecker.
type RevenueCatConfig struct {
	// APIKey is a scoped RevenueCat v2 secret API key. It only needs
	// customer_information:customers:read permission.
	APIKey string
	// ProjectID is the RevenueCat project ID containing the customer and
	// entitlement configuration.
	ProjectID string
	// EntitlementID is the RevenueCat entitlement ID (the entl... identifier,
	// not its lookup key) that grants access.
	EntitlementID string
	// CacheTTL is how long a positive check result is cached. Defaults to
	// DefaultCacheTTL.
	CacheTTL time.Duration
	// NegativeCacheTTL is how long a negative check result is cached.
	// Defaults to DefaultNegativeCacheTTL.
	NegativeCacheTTL time.Duration
	// BaseURL overrides the RevenueCat API base URL. Used in tests.
	BaseURL string
}

type cacheEntry struct {
	active    bool
	expiresAt time.Time
}

// RevenueCatChecker checks subscriptions against the RevenueCat REST API v2.
// The Keycloak user ID is used as the RevenueCat App User ID on every platform,
// so a purchase made in the web client (Paddle) or in the mobile app (App
// Store / Google Play) unlocks the same account.
//
// Results are cached per user for the configured TTL to keep request latency
// low and stay well below RevenueCat's rate limits. Concurrent checks for the
// same user are collapsed into a single API call (single-flight) so that
// parallel requests after a cache expiry don't cause a stampede.
type RevenueCatChecker struct {
	baseURL       string
	apiKey        string
	projectID     string
	entitlementID string
	httpClient       *http.Client
	cacheTTL         time.Duration
	negativeCacheTTL time.Duration
	now              func() time.Time

	inFlight singleflight.Group

	cacheMutex sync.Mutex
	cache      map[string]cacheEntry // keyed by Keycloak user ID
}

func NewRevenueCatChecker(cfg RevenueCatConfig) *RevenueCatChecker {
	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	negativeTTL := cfg.NegativeCacheTTL
	if negativeTTL <= 0 {
		negativeTTL = DefaultNegativeCacheTTL
	}
	if negativeTTL > ttl {
		negativeTTL = ttl
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = revenueCatBaseURL
	}
	return &RevenueCatChecker{
		baseURL:          baseURL,
		apiKey:           cfg.APIKey,
		projectID:        cfg.ProjectID,
		entitlementID:    cfg.EntitlementID,
		httpClient:       &http.Client{Timeout: 10 * time.Second},
		cacheTTL:         ttl,
		negativeCacheTTL: negativeTTL,
		now:              time.Now,
		cache:            make(map[string]cacheEntry),
	}
}

func (c *RevenueCatChecker) IsEnabled() bool {
	return true
}

func (c *RevenueCatChecker) CheckSubscription(ctx context.Context, userID string) (bool, error) {
	if entry, ok := c.getCached(userID); ok {
		slog.Debug("Using cached subscription status", "userID", userID, "active", entry.active)
		return entry.active, nil
	}
	entry, err := c.resolveWithSingleFlight(ctx, userID)
	if err != nil {
		return false, err
	}
	return entry.active, nil
}

func (c *RevenueCatChecker) RefreshSubscription(ctx context.Context, userID string) (bool, error) {
	c.clearUserCache(userID)
	return c.CheckSubscription(ctx, userID)
}

// ClearCache drops all cached subscription states.
func (c *RevenueCatChecker) ClearCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache = make(map[string]cacheEntry)
}

// resolveWithSingleFlight resolves the user's subscription state, collapsing
// concurrent resolutions for the same user into one RevenueCat round trip. The
// first caller's context drives the shared call; if it is cancelled the other
// waiters receive the error and can retry.
func (c *RevenueCatChecker) resolveWithSingleFlight(ctx context.Context, userID string) (cacheEntry, error) {
	result, err, _ := c.inFlight.Do(userID, func() (any, error) {
		return c.resolveAndCache(ctx, userID)
	})
	if err != nil {
		return cacheEntry{}, err
	}
	entry, ok := result.(cacheEntry)
	if !ok {
		return cacheEntry{}, fmt.Errorf("unexpected revenuecat resolution result type %T", result)
	}
	return entry, nil
}

// resolveAndCache fetches the user's customer record from RevenueCat and
// checks the configured entitlement in its active_entitlements list.
func (c *RevenueCatChecker) resolveAndCache(ctx context.Context, userID string) (cacheEntry, error) {
	var response revenueCatCustomerResponse
	path := "/projects/" + url.PathEscape(c.projectID) + "/customers/" + url.PathEscape(userID)
	status, err := c.doJSON(ctx, http.MethodGet, path, &response)
	if err != nil {
		return cacheEntry{}, err
	}

	var entry cacheEntry
	switch status {
	case http.StatusOK:
		entry = cacheEntry{
			active: c.hasActiveEntitlement(response.Customer.ActiveEntitlements.Items),
		}
		slog.Info("RevenueCat subscription check completed", "userID", userID, "active", entry.active)
		slog.Debug("RevenueCat active entitlements", "userID", userID, "entitlement_ids", activeEntitlementIDs(response.Customer.ActiveEntitlements.Items))
	case http.StatusNotFound:
		// RevenueCat has never seen this user: no purchase, definitive answer.
		entry = cacheEntry{active: false}
		slog.Info("No RevenueCat customer found for user, subscription inactive", "userID", userID)
	default:
		return cacheEntry{}, fmt.Errorf("revenuecat customer lookup failed with status %d", status)
	}

	entry.expiresAt = c.now().Add(c.cacheTTLFor(entry.active))
	c.cacheUser(userID, entry)
	return entry, nil
}

// hasActiveEntitlement reports whether the configured entitlement ID appears
// in RevenueCat's active_entitlements list. The API returns only active
// entitlements; expires_at is a Unix timestamp in milliseconds and can be null
// for an entitlement without a fixed expiry.
func (c *RevenueCatChecker) hasActiveEntitlement(entitlements []revenueCatActiveEntitlement) bool {
	for _, entitlement := range entitlements {
		if entitlement.EntitlementID != c.entitlementID {
			continue
		}
		if entitlement.ExpiresAt == nil {
			return true
		}
		if c.now().Before(time.UnixMilli(*entitlement.ExpiresAt)) {
			return true
		}
	}
	return false
}

// activeEntitlementIDs lists the entitlement IDs RevenueCat reports as active
// (for debugging config mismatches: the configured entitlement must be the
// internal ID from the dashboard, not the human-readable lookup key).
func activeEntitlementIDs(entitlements []revenueCatActiveEntitlement) []string {
	ids := make([]string, 0, len(entitlements))
	for _, entitlement := range entitlements {
		ids = append(ids, entitlement.EntitlementID)
	}
	return ids
}

// cacheTTLFor returns the cache duration for a check result. Negative results
// expire much faster so a just-purchased subscription is not blocked by a
// stale 402.
func (c *RevenueCatChecker) cacheTTLFor(active bool) time.Duration {
	if active {
		return c.cacheTTL
	}
	return c.negativeCacheTTL
}

func (c *RevenueCatChecker) getCached(userID string) (cacheEntry, bool) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	entry, ok := c.cache[userID]
	if !ok || c.now().After(entry.expiresAt) {
		return cacheEntry{}, false
	}
	return entry, true
}

func (c *RevenueCatChecker) cacheUser(userID string, entry cacheEntry) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache[userID] = entry
}

func (c *RevenueCatChecker) clearUserCache(userID string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.cache, userID)
}

// doJSON performs an authenticated RevenueCat API request. It returns the HTTP
// status code; a non-nil error means a transport-level failure so callers can
// fail open.
func (c *RevenueCatChecker) doJSON(ctx context.Context, method, path string, out any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create revenuecat request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("revenuecat request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("failed to read revenuecat response: %w", err)
	}
	// Only 2xx responses carry a payload we care about; error responses are
	// evaluated by their status code alone, so an HTML error page from a proxy
	// must not turn into a decode error.
	if out != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return resp.StatusCode, fmt.Errorf("failed to decode revenuecat response (status %d, body %.120q): %w", resp.StatusCode, data, err)
		}
	}
	return resp.StatusCode, nil
}

type revenueCatCustomerResponse struct {
	Customer revenueCatCustomer `json:"-"`
}

// The v2 get-customer response is the customer object itself, not wrapped in
// a "customer" property.
func (r *revenueCatCustomerResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Customer)
}

type revenueCatCustomer struct {
	ActiveEntitlements revenueCatActiveEntitlementList `json:"active_entitlements"`
}

type revenueCatActiveEntitlementList struct {
	Items []revenueCatActiveEntitlement `json:"items"`
}

type revenueCatActiveEntitlement struct {
	EntitlementID string `json:"entitlement_id"`
	ExpiresAt     *int64 `json:"expires_at"`
}
