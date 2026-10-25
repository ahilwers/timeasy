package subscription

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	// PolarEnvSandbox selects Polar's sandbox environment (test cards, no real money).
	PolarEnvSandbox = "sandbox"
	// PolarEnvProduction selects Polar's production environment.
	PolarEnvProduction = "production"
)

// DefaultCacheTTL is the default duration a subscription check result stays valid.
const DefaultCacheTTL = 15 * time.Minute

// PolarConfig holds the configuration for the PolarChecker.
type PolarConfig struct {
	// Env is either PolarEnvSandbox or PolarEnvProduction.
	Env string
	// OrganizationID is the Polar organization ID.
	OrganizationID string
	// ProductID is the subscription product users must have active. When empty,
	// any active subscription grants access.
	ProductID string
	// AccessToken is the Polar organization access token.
	AccessToken string
	// CacheTTL is how long a check result is cached. Defaults to DefaultCacheTTL.
	CacheTTL time.Duration
}

type cacheEntry struct {
	active          bool
	polarCustomerID string
	expiresAt       time.Time
}

// PolarChecker checks subscriptions directly against the Polar API.
// Results are cached per user for the configured TTL to keep request latency
// low and stay well below Polar's rate limits. Concurrent checks for the same
// user are collapsed into a single Polar call (single-flight) so that parallel
// requests after a cache expiry don't cause a stampede.
type PolarChecker struct {
	baseURL     string
	accessToken string
	productID   string
	httpClient  *http.Client
	cacheTTL    time.Duration
	now         func() time.Time

	inFlight singleflight.Group

	cacheMutex sync.Mutex
	cache      map[string]cacheEntry // keyed by Keycloak user ID
}

func NewPolarChecker(cfg PolarConfig) *PolarChecker {
	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	return &PolarChecker{
		baseURL:     polarBaseURL(cfg.Env),
		accessToken: cfg.AccessToken,
		productID:   cfg.ProductID,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		cacheTTL:    ttl,
		now:         time.Now,
		cache:       make(map[string]cacheEntry),
	}
}

// localePattern matches BCP 47 language tags like "de", "de-DE" or "en-GB".
var localePattern = regexp.MustCompile(`^[A-Za-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$`)

// normalizeLocale returns the locale if it is a plausible BCP 47 language tag,
// otherwise "". Invalid values are omitted so Polar falls back to English
// instead of rejecting the checkout.
func normalizeLocale(locale string) string {
	locale = strings.TrimSpace(locale)
	if locale == "" || len(locale) > 35 || !localePattern.MatchString(locale) {
		return ""
	}
	return locale
}

// polarBaseURL returns the Polar API base URL for the configured environment.
func polarBaseURL(env string) string {
	if strings.EqualFold(env, PolarEnvProduction) {
		return "https://api.polar.sh"
	}
	return "https://sandbox-api.polar.sh"
}

func (c *PolarChecker) IsEnabled() bool {
	return true
}

func (c *PolarChecker) CheckSubscription(ctx context.Context, userID, email string) (bool, error) {
	if entry, ok := c.getCached(userID); ok {
		slog.Debug("Using cached subscription status", "userID", userID, "active", entry.active)
		return entry.active, nil
	}
	entry, err := c.resolveWithSingleFlight(ctx, userID, email)
	if err != nil {
		return false, err
	}
	return entry.active, nil
}

func (c *PolarChecker) RefreshSubscription(ctx context.Context, userID, email string) (bool, error) {
	c.clearUserCache(userID)
	return c.CheckSubscription(ctx, userID, email)
}

func (c *PolarChecker) GetPolarCustomerID(ctx context.Context, userID, email string) (string, error) {
	if entry, ok := c.getCached(userID); ok {
		return entry.polarCustomerID, nil
	}
	entry, err := c.resolveWithSingleFlight(ctx, userID, email)
	if err != nil {
		return "", err
	}
	return entry.polarCustomerID, nil
}

// resolveWithSingleFlight resolves the user's subscription state, collapsing
// concurrent resolutions for the same user into one Polar round trip. The
// first caller's context drives the shared call; if it is cancelled the other
// waiters receive the error and can retry.
func (c *PolarChecker) resolveWithSingleFlight(ctx context.Context, userID, email string) (cacheEntry, error) {
	result, err, _ := c.inFlight.Do(userID, func() (any, error) {
		return c.resolveAndCache(ctx, userID, email)
	})
	if err != nil {
		return cacheEntry{}, err
	}
	entry, ok := result.(cacheEntry)
	if !ok {
		return cacheEntry{}, fmt.Errorf("unexpected polar resolution result type %T", result)
	}
	return entry, nil
}

// CreateCheckoutSession creates a Polar hosted checkout session for the user
// and returns its URL. The session links the created Polar customer to the
// Keycloak user via external_customer_id so future lookups are direct.
// locale is an optional BCP 47 tag for the checkout language; unsupported or
// empty values fall back to Polar's default (English).
func (c *PolarChecker) CreateCheckoutSession(ctx context.Context, userID, email, successURL, locale string) (string, error) {
	body := map[string]any{
		"products":             []string{c.productID},
		"external_customer_id": userID,
	}
	if email != "" {
		body["customer_email"] = email
	}
	if successURL != "" {
		body["success_url"] = successURL
	}
	if normalized := normalizeLocale(locale); normalized != "" {
		body["locale"] = normalized
	}
	var session polarCheckoutSession
	status, err := c.doJSON(ctx, http.MethodPost, "/v1/checkouts/", body, &session)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return "", fmt.Errorf("polar checkout creation failed with status %d", status)
	}
	return session.URL, nil
}

// CreatePortalSession creates a short-lived Polar customer portal session for
// the user and returns the portal URL, or "" when the user has no Polar
// customer yet.
//
// Note: the custom portal has no locale parameter and Polar's hosted portal is
// English-only (rendered with lang="en"), so the customer's locale cannot
// localize it.
func (c *PolarChecker) CreatePortalSession(ctx context.Context, userID, email string) (string, error) {
	customerID, err := c.GetPolarCustomerID(ctx, userID, email)
	if err != nil {
		return "", err
	}
	if customerID == "" {
		return "", nil
	}
	var session polarCustomerSession
	status, err := c.doJSON(ctx, http.MethodPost, "/v1/customer-sessions/",
		map[string]any{"customer_id": customerID}, &session)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return "", fmt.Errorf("polar customer session creation failed with status %d", status)
	}
	return session.CustomerPortalURL, nil
}

// ClearCache drops all cached subscription states.
func (c *PolarChecker) ClearCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache = make(map[string]cacheEntry)
}

// resolveAndCache resolves the user's Polar customer and subscription state
// and caches the result.
func (c *PolarChecker) resolveAndCache(ctx context.Context, userID, email string) (cacheEntry, error) {
	customerID, err := c.resolveCustomer(ctx, userID, email)
	if err != nil {
		return cacheEntry{}, err
	}
	if customerID == "" {
		// Definitive: no Polar customer exists for this user.
		slog.Info("No Polar customer found for user, subscription inactive", "userID", userID)
		entry := cacheEntry{active: false, expiresAt: c.now().Add(c.cacheTTL)}
		c.cacheUser(userID, entry)
		return entry, nil
	}
	state, err := c.fetchCustomerState(ctx, customerID)
	if err != nil {
		return cacheEntry{}, err
	}
	entry := cacheEntry{
		active:          c.hasActiveSubscription(state),
		polarCustomerID: customerID,
		expiresAt:       c.now().Add(c.cacheTTL),
	}
	c.cacheUser(userID, entry)
	slog.Info("Polar subscription check completed", "userID", userID, "active", entry.active)
	return entry, nil
}

// resolveCustomer finds the Polar customer for a Keycloak user. It first looks
// the customer up by external ID (= Keycloak user ID, set when the checkout
// was created by our server). If none exists, it falls back to matching by
// email and backfills the external ID so later lookups are direct. Returns ""
// when no customer exists.
func (c *PolarChecker) resolveCustomer(ctx context.Context, userID, email string) (string, error) {
	customer, status, err := c.getCustomerByExternalID(ctx, userID)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		// 404 means the Polar customer for this user doesn't exist yet.
		customer, err = c.findCustomerByEmail(ctx, email)
		if err != nil {
			return "", err
		}
		if customer == nil {
			return "", nil
		}
		// Backfill the external ID so future lookups are direct. Best effort:
		// failures are logged but don't block (the customer ID is known anyway).
		if err := c.backfillExternalID(ctx, customer.ID, userID); err != nil {
			slog.Warn("Failed to backfill Polar external_id", "customerID", customer.ID, "error", err)
		}
	}
	return customer.ID, nil
}

func (c *PolarChecker) getCustomerByExternalID(ctx context.Context, externalID string) (*polarCustomer, int, error) {
	var customer polarCustomer
	status, err := c.doJSON(ctx, http.MethodGet,
		fmt.Sprintf("/v1/customers/external/%s", url.PathEscape(externalID)), nil, &customer)
	if err != nil {
		return nil, 0, err
	}
	if status != http.StatusOK {
		return nil, status, nil
	}
	return &customer, status, nil
}

// findCustomerByEmail looks up a Polar customer by exact email match. Returns
// nil when no matching customer exists.
func (c *PolarChecker) findCustomerByEmail(ctx context.Context, email string) (*polarCustomer, error) {
	if email == "" {
		return nil, nil
	}
	var list polarCustomerList
	status, err := c.doJSON(ctx, http.MethodGet,
		fmt.Sprintf("/v1/customers?email=%s&limit=10", url.QueryEscape(email)), nil, &list)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("polar customer lookup failed with status %d", status)
	}
	for i := range list.Items {
		if list.Items[i].Email != nil && strings.EqualFold(*list.Items[i].Email, email) {
			customer := list.Items[i]
			return &customer, nil
		}
	}
	return nil, nil
}

func (c *PolarChecker) backfillExternalID(ctx context.Context, customerID, externalID string) error {
	var out map[string]any
	status, err := c.doJSON(ctx, http.MethodPatch,
		fmt.Sprintf("/v1/customers/%s", customerID),
		map[string]string{"external_id": externalID}, &out)
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("polar external_id backfill failed with status %d", status)
	}
	return nil
}

func (c *PolarChecker) fetchCustomerState(ctx context.Context, customerID string) (*polarCustomerState, error) {
	var state polarCustomerState
	status, err := c.doJSON(ctx, http.MethodGet,
		fmt.Sprintf("/v1/customers/%s/state", customerID), nil, &state)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("polar customer state lookup failed with status %d", status)
	}
	return &state, nil
}

// hasActiveSubscription reports whether the customer state contains an active
// subscription for the configured product (any product when none is
// configured). The state endpoint only lists active/trialing subscriptions; a
// subscription canceled at period end stays listed until the period ends,
// which matches the access we want to grant.
func (c *PolarChecker) hasActiveSubscription(state *polarCustomerState) bool {
	for _, sub := range state.ActiveSubscriptions {
		if c.productID == "" || sub.ProductID == c.productID {
			return true
		}
	}
	return false
}

func (c *PolarChecker) getCached(userID string) (cacheEntry, bool) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	entry, ok := c.cache[userID]
	if !ok || c.now().After(entry.expiresAt) {
		return cacheEntry{}, false
	}
	return entry, true
}

func (c *PolarChecker) cacheUser(userID string, entry cacheEntry) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache[userID] = entry
}

func (c *PolarChecker) clearUserCache(userID string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.cache, userID)
}

// doJSON performs an authenticated Polar API request. It returns the HTTP
// status code; a non-nil error means a transport-level failure so callers can
// fail open.
func (c *PolarChecker) doJSON(ctx context.Context, method, path string, body any, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal polar request: %w", err)
		}
		reader = bytes.NewBuffer(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, fmt.Errorf("failed to create polar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("polar request failed: %w", err)
	}
	defer resp.Body.Close()
	if out != nil {
		data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			return resp.StatusCode, fmt.Errorf("failed to read polar response: %w", err)
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, out); err != nil {
				return resp.StatusCode, fmt.Errorf("failed to decode polar response: %w", err)
			}
		}
	}
	return resp.StatusCode, nil
}

type polarCustomer struct {
	ID    string  `json:"id"`
	Email *string `json:"email"`
}

type polarCustomerList struct {
	Items []polarCustomer `json:"items"`
}

type polarSubscription struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	ProductID string `json:"product_id"`
}

type polarCustomerState struct {
	ID                  string              `json:"id"`
	ActiveSubscriptions []polarSubscription `json:"active_subscriptions"`
}

type polarCheckoutSession struct {
	URL string `json:"url"`
}

type polarCustomerSession struct {
	CustomerPortalURL string `json:"customer_portal_url"`
}
