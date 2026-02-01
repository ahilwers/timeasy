package license

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// TokenRequest represents the authentication request to the license manager
type TokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// TokenResponse represents the authentication response from the license manager
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type CheckSubscriptionResponse struct {
	Active bool `json:"active"`
}

type cachedToken struct {
	accessToken string
	expiresAt   time.Time
}

type cachedSubscription struct {
	active    bool
	expiresAt time.Time
}

// LicenseManagerClient handles communication with the license manager service
type LicenseManagerClient struct {
	host         string
	clientID     string
	clientSecret string
	productID    string
	httpClient   *http.Client

	tokenMutex sync.RWMutex
	token      *cachedToken

	subscriptionMutex sync.RWMutex
	subscriptions     map[string]*cachedSubscription // keyed by userID
}

func NewLicenseManagerClient(host, clientID, clientSecret, productID string) *LicenseManagerClient {
	return &LicenseManagerClient{
		host:          host,
		clientID:      clientID,
		clientSecret:  clientSecret,
		productID:     productID,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		subscriptions: make(map[string]*cachedSubscription),
	}
}

// IsConfigured returns true if the license manager host is defined
// When not configured, license checks are skipped and access is always allowed
func (c *LicenseManagerClient) IsConfigured() bool {
	return c.host != ""
}

// CheckSubscription checks if a user has an active subscription
// Positive results are cached for 24 hours per user
func (c *LicenseManagerClient) CheckSubscription(userID string) (bool, error) {
	if !c.IsConfigured() {
		slog.Debug("License manager not configured, allowing access")
		return true, nil
	}

	if cached := c.getCachedSubscription(userID); cached != nil {
		slog.Debug("Using cached subscription status", "userID", userID, "active", cached.active)
		return cached.active, nil
	}

	token, err := c.getToken()
	if err != nil {
		return false, fmt.Errorf("failed to get license manager token: %w", err)
	}

	active, err := c.checkSubscriptionWithAPI(userID, token)
	if err != nil {
		return false, fmt.Errorf("failed to check subscription: %w", err)
	}

	if active {
		c.cacheSubscription(userID, active)
	}

	slog.Info("Subscription check completed", "userID", userID, "active", active)
	return active, nil
}

// getToken returns a valid authentication token, fetching a new one if necessary
func (c *LicenseManagerClient) getToken() (string, error) {
	c.tokenMutex.RLock()
	if c.token != nil && time.Now().Before(c.token.expiresAt) {
		token := c.token.accessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	// Need to fetch a new token
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	// Double-check after acquiring write lock
	if c.token != nil && time.Now().Before(c.token.expiresAt) {
		return c.token.accessToken, nil
	}

	tokenResp, err := c.authenticate()
	if err != nil {
		return "", err
	}

	// Cache token with some buffer before expiry (5 minutes)
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - 5*time.Minute)
	c.token = &cachedToken{
		accessToken: tokenResp.AccessToken,
		expiresAt:   expiresAt,
	}

	return tokenResp.AccessToken, nil
}

// authenticate performs authentication with the license manager
func (c *LicenseManagerClient) authenticate() (*TokenResponse, error) {
	url := fmt.Sprintf("%s/api/auth/token", c.host)

	reqBody := TokenRequest{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// checkSubscriptionWithAPI calls the license manager API to check subscription status
func (c *LicenseManagerClient) checkSubscriptionWithAPI(userID, token string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/subscriptions/check?userId=%s&productId=%s", c.host, userID, c.productID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create subscription check request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send subscription check request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("subscription check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var checkResp CheckSubscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return false, fmt.Errorf("failed to decode subscription check response: %w", err)
	}

	return checkResp.Active, nil
}

func (c *LicenseManagerClient) getCachedSubscription(userID string) *cachedSubscription {
	c.subscriptionMutex.RLock()
	defer c.subscriptionMutex.RUnlock()

	cached, ok := c.subscriptions[userID]
	if !ok || time.Now().After(cached.expiresAt) {
		return nil
	}
	return cached
}

func (c *LicenseManagerClient) cacheSubscription(userID string, active bool) {
	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()

	c.subscriptions[userID] = &cachedSubscription{
		active:    active,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
}

func (c *LicenseManagerClient) ClearCache() {
	c.tokenMutex.Lock()
	c.token = nil
	c.tokenMutex.Unlock()

	c.subscriptionMutex.Lock()
	c.subscriptions = make(map[string]*cachedSubscription)
	c.subscriptionMutex.Unlock()
}

func (c *LicenseManagerClient) ClearUserCache(userID string) {
	c.subscriptionMutex.Lock()
	defer c.subscriptionMutex.Unlock()
	delete(c.subscriptions, userID)
}
