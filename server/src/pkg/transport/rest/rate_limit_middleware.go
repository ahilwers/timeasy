package rest

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitMiddleware defines the interface for IP-based rate limiting middleware.
type RateLimitMiddleware interface {
	HandlerFunc() gin.HandlerFunc
}

const (
	// How often stale client buckets are removed from memory.
	rateLimitCleanupInterval = time.Minute
	// A client that has not sent a request for this long loses its bucket
	// and starts with a fresh one on the next request.
	rateLimitClientExpiry = 10 * time.Minute
)

type rateLimitMiddleware struct {
	requestsPerMinute int
	burst             int

	mu       sync.Mutex
	limiters map[string]*rateLimiterEntry
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimitMiddleware creates IP-based rate limiting middleware. Each
// client IP gets its own token bucket: requestsPerMinute is the sustained
// rate, burst allows short-term spikes above that rate. Exceeding the limit
// answers with 429 Too Many Requests and a Retry-After header.
func NewRateLimitMiddleware(requestsPerMinute int, burst int) RateLimitMiddleware {
	mw := &rateLimitMiddleware{
		requestsPerMinute: requestsPerMinute,
		burst:             burst,
		limiters:          make(map[string]*rateLimiterEntry),
	}
	go mw.cleanupLoop()
	return mw
}

// NewNoopRateLimitMiddleware creates a rate limiting middleware that allows
// every request. Used when rate limiting is disabled via configuration.
func NewNoopRateLimitMiddleware() RateLimitMiddleware {
	return &noopRateLimitMiddleware{}
}

type noopRateLimitMiddleware struct{}

func (mw *noopRateLimitMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (mw *rateLimitMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		entry := mw.getLimiter(key)

		allowed, wait := entry.tryTake()
		if !allowed {
			retryAfterSeconds := int(math.Ceil(wait.Seconds()))
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}
			slog.Warn("Rate limit exceeded",
				"client_ip", key,
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"retry_after_seconds", retryAfterSeconds)
			c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
			c.Header("X-RateLimit-Limit", strconv.Itoa(mw.requestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(wait).Unix(), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many requests. Please slow down and retry later.",
			})
			return
		}

		mw.setRateLimitHeaders(c, entry)
		c.Next()
	}
}

// tryTake reports whether a request can be served immediately. When it
// returns false, wait is the time until the next token becomes available.
// The reservation pattern is used instead of Allow() so the exact wait time
// can be reported; rejected requests cancel their reservation again so they
// do not consume future capacity.
func (entry *rateLimiterEntry) tryTake() (allowed bool, wait time.Duration) {
	reservation := entry.limiter.Reserve()
	if !reservation.OK() {
		// Only possible when a single token exceeds the burst size, which
		// valid configurations never do. Be defensive anyway.
		return false, time.Second
	}
	delay := reservation.Delay()
	if delay > 0 {
		reservation.Cancel()
		return false, delay
	}
	return true, 0
}

// setRateLimitHeaders advertises the bucket state to the client: the
// sustained limit, the tokens left after this request and the time (unix
// seconds) when the bucket will be full again.
func (mw *rateLimitMiddleware) setRateLimitHeaders(c *gin.Context, entry *rateLimiterEntry) {
	tokens := entry.limiter.Tokens()
	remaining := int(math.Floor(tokens))
	if remaining < 0 {
		remaining = 0
	}
	missing := float64(entry.limiter.Burst()) - tokens
	if missing < 0 {
		missing = 0
	}
	reset := time.Now().Add(time.Duration(missing / float64(entry.limiter.Limit()) * float64(time.Second)))
	c.Header("X-RateLimit-Limit", strconv.Itoa(mw.requestsPerMinute))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
}

func (mw *rateLimitMiddleware) getLimiter(key string) *rateLimiterEntry {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	entry, ok := mw.limiters[key]
	if !ok {
		entry = &rateLimiterEntry{
			limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(mw.requestsPerMinute)), mw.burst),
		}
		mw.limiters[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry
}

// cleanupLoop periodically drops buckets of clients that have been idle for
// longer than rateLimitClientExpiry so the map cannot grow without bounds.
func (mw *rateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(rateLimitCleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		mw.cleanup()
	}
}

func (mw *rateLimitMiddleware) cleanup() {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	for key, entry := range mw.limiters {
		if time.Since(entry.lastSeen) > rateLimitClientExpiry {
			delete(mw.limiters, key)
		}
	}
}
