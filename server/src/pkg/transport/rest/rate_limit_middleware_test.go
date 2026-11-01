package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newRateLimitTestEngine(requestsPerMinute, burst int, trustedProxies []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		panic(err)
	}
	router.Use(NewRateLimitMiddleware(requestsPerMinute, burst).HandlerFunc())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return router
}

func performRateLimitRequest(router *gin.Engine, remoteAddr string, forwardedFor string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestRateLimitMiddlewareAllowsRequestsWithinLimit(t *testing.T) {
	router := newRateLimitTestEngine(6000, 10, nil)

	for i := 0; i < 10; i++ {
		response := performRateLimitRequest(router, "192.0.2.1:1234", "")
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "6000", response.Header().Get("X-RateLimit-Limit"))
		require.NotEmpty(t, response.Header().Get("X-RateLimit-Remaining"))
		require.NotEmpty(t, response.Header().Get("X-RateLimit-Reset"))
	}
}

func TestRateLimitMiddlewareRejectsRequestsAboveLimit(t *testing.T) {
	// One request per second with a small burst: the third request within the
	// same second must be rejected.
	router := newRateLimitTestEngine(60, 2, nil)

	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "192.0.2.1:1234", "").Code)
	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "192.0.2.1:1234", "").Code)

	rejected := performRateLimitRequest(router, "192.0.2.1:1234", "")
	require.Equal(t, http.StatusTooManyRequests, rejected.Code)
	require.Equal(t, "0", rejected.Header().Get("X-RateLimit-Remaining"))

	retryAfter := rejected.Header().Get("Retry-After")
	require.NotEmpty(t, retryAfter)
	retryAfterSeconds, err := strconv.Atoi(retryAfter)
	require.NoError(t, err)
	require.GreaterOrEqual(t, retryAfterSeconds, 1)

	reset, err := strconv.ParseInt(rejected.Header().Get("X-RateLimit-Reset"), 10, 64)
	require.NoError(t, err)
	require.Greater(t, reset, time.Now().Unix())

	var body map[string]string
	require.NoError(t, json.Unmarshal(rejected.Body.Bytes(), &body))
	require.Equal(t, "rate_limited", body["error"])
}

func TestRateLimitMiddlewareKeepsSeparateBucketsPerClient(t *testing.T) {
	router := newRateLimitTestEngine(60, 1, nil)

	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "192.0.2.1:1234", "").Code)
	require.Equal(t, http.StatusTooManyRequests, performRateLimitRequest(router, "192.0.2.1:1234", "").Code)
	// A different client still has its own fresh bucket.
	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "192.0.2.2:1234", "").Code)
}

func TestRateLimitMiddlewareUsesForwardedForFromTrustedProxy(t *testing.T) {
	router := newRateLimitTestEngine(60, 1, []string{"0.0.0.0/0", "::/0"})

	// Both requests come from the same proxy but different real clients, so
	// they must not share a bucket.
	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "10.0.0.1:1234", "203.0.113.7").Code)
	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "10.0.0.1:1234", "203.0.113.8").Code)
	// The same real client is limited independently.
	require.Equal(t, http.StatusTooManyRequests, performRateLimitRequest(router, "10.0.0.1:1234", "203.0.113.7").Code)
}

func TestRateLimitMiddlewareIgnoresForwardedForFromUntrustedProxy(t *testing.T) {
	router := newRateLimitTestEngine(60, 1, nil)

	// Without trusting the proxy, the forwarded header must not create new
	// buckets: both requests count against the proxy's own IP.
	require.Equal(t, http.StatusOK, performRateLimitRequest(router, "10.0.0.1:1234", "203.0.113.7").Code)
	require.Equal(t, http.StatusTooManyRequests, performRateLimitRequest(router, "10.0.0.1:1234", "203.0.113.8").Code)
}

func TestNoopRateLimitMiddlewareAllowsEverything(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewNoopRateLimitMiddleware().HandlerFunc())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 100; i++ {
		response := performRateLimitRequest(router, "192.0.2.1:1234", "")
		require.Equal(t, http.StatusOK, response.Code)
		require.Empty(t, response.Header().Get("X-RateLimit-Limit"))
	}
}
