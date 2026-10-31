package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

const (
	jwksCacheTTL           = 15 * time.Minute
	jwksMinRefreshInterval = time.Minute
	jwksMaxSizeBytes       = 1 << 20 // 1 MiB
	jwksFetchTimeout       = 10 * time.Second
)

type keycloakTokenVerifier struct {
	keycloakHost string
	realm        string
	httpClient   *http.Client

	jwksMu             sync.RWMutex
	cachedJwks         jwk.Set
	cachedJwksAt       time.Time
	lastRefreshAttempt time.Time
}

func NewKeycloakTokenVerifier(keycloakHost string, realm string) TokenVerifier {
	return &keycloakTokenVerifier{
		keycloakHost: keycloakHost,
		realm:        realm,
		httpClient:   &http.Client{Timeout: jwksFetchTimeout},
	}
}

func (v *keycloakTokenVerifier) VerifyToken(c *gin.Context) (AuthToken, error) {
	strToken, err := v.getAuthHeader(c)
	if err != nil {
		return nil, err
	}
	jwksKeySet, err := v.getJwks(c.Request)
	if err != nil {
		return nil, err
	}
	token, err := jwt.Parse([]byte(strToken), jwt.WithKeySet(jwksKeySet))
	if err != nil {
		// The signing key may have been rotated, so refresh the key set once
		// and retry before rejecting the request.
		refreshedKeySet, refreshErr := v.refreshJwks(c.Request.Context())
		if refreshErr != nil {
			return nil, err
		}
		token, err = jwt.Parse([]byte(strToken), jwt.WithKeySet(refreshedKeySet))
		if err != nil {
			return nil, err
		}
	}
	return NewKeycloakToken(token), nil
}

// getJwks returns the cached key set if it is still fresh, otherwise it
// fetches it again from Keycloak.
func (v *keycloakTokenVerifier) getJwks(req *http.Request) (jwk.Set, error) {
	v.jwksMu.RLock()
	cached := v.cachedJwks
	cachedAt := v.cachedJwksAt
	v.jwksMu.RUnlock()

	if cached != nil && time.Since(cachedAt) < jwksCacheTTL {
		return cached, nil
	}

	fetched, err := v.refreshJwks(req.Context())
	if err != nil {
		// Fall back to a stale key set instead of rejecting every request
		// while Keycloak is temporarily unavailable.
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}
	return fetched, nil
}

// refreshJwks fetches the key set from Keycloak and updates the cache. Refresh
// attempts are rate limited so that requests with invalid tokens cannot trigger
// a fetch from Keycloak every time.
func (v *keycloakTokenVerifier) refreshJwks(ctx context.Context) (jwk.Set, error) {
	v.jwksMu.Lock()
	throttled := !v.lastRefreshAttempt.IsZero() && time.Since(v.lastRefreshAttempt) < jwksMinRefreshInterval
	if !throttled {
		v.lastRefreshAttempt = time.Now()
	}
	cached := v.cachedJwks
	v.jwksMu.Unlock()

	if throttled && cached != nil {
		return cached, nil
	}

	fetched, err := v.fetchJwks(ctx)
	if err != nil {
		return nil, err
	}

	v.jwksMu.Lock()
	v.cachedJwks = fetched
	v.cachedJwksAt = time.Now()
	v.jwksMu.Unlock()

	return fetched, nil
}

func (v *keycloakTokenVerifier) fetchJwks(ctx context.Context) (jwk.Set, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.getKeycloakJwksUrl(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, jwksMaxSizeBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	keySet, err := jwk.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}
	return keySet, nil
}

func (v *keycloakTokenVerifier) getAuthHeader(c *gin.Context) (string, error) {
	header := strings.Fields(c.Request.Header.Get("Authorization"))
	if len(header) < 2 {
		return "", fmt.Errorf("authorization header missing")
	}
	if header[0] != "Bearer" {
		return "", fmt.Errorf("malformed token")
	}
	return header[1], nil
}

func (v *keycloakTokenVerifier) getKeycloakJwksUrl() string {
	url := fmt.Sprintf("%v/realms/%v/protocol/openid-connect/certs", v.keycloakHost, v.realm)
	return url
}
