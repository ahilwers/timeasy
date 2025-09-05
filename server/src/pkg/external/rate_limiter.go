package external

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProviderRateLimiter manages rate limiting for external API providers
type ProviderRateLimiter struct {
	mu         sync.RWMutex
	providers  map[string]*ProviderLimitState
	rateLimits map[string]RateLimitConfig
}

// RateLimitConfig defines rate limiting parameters for a provider
type RateLimitConfig struct {
	RequestsPerWindow int           // Number of requests allowed per window
	WindowSize        time.Duration // Size of the rate limiting window
	MinInterval       time.Duration // Minimum time between requests
	BackoffBase       time.Duration // Base time for exponential backoff
	MaxBackoff        time.Duration // Maximum backoff time
}

// ProviderLimitState tracks the current rate limiting state for a provider
type ProviderLimitState struct {
	lastRequest      time.Time
	requestCount     int
	windowStart      time.Time
	currentBackoff   time.Duration
	consecutiveFails int
}

func NewProviderRateLimiter() *ProviderRateLimiter {
	limiter := &ProviderRateLimiter{
		providers:  make(map[string]*ProviderLimitState),
		rateLimits: make(map[string]RateLimitConfig),
	}

	// Configure default rate limits for each provider
	limiter.rateLimits["github"] = RateLimitConfig{
		RequestsPerWindow: 30, // Conservative limit (GitHub allows 5000/hour = ~83/min)
		WindowSize:        time.Minute,
		MinInterval:       2 * time.Second, // ~30 requests per minute
		BackoffBase:       10 * time.Second,
		MaxBackoff:        10 * time.Minute,
	}

	limiter.rateLimits["gitlab"] = RateLimitConfig{
		RequestsPerWindow: 30, // Conservative limit (GitLab allows 2000/hour = ~33/min)
		WindowSize:        time.Minute,
		MinInterval:       2 * time.Second, // ~30 requests per minute
		BackoffBase:       10 * time.Second,
		MaxBackoff:        10 * time.Minute,
	}

	limiter.rateLimits["jira"] = RateLimitConfig{
		RequestsPerWindow: 10, // Very conservative (Jira limits vary widely)
		WindowSize:        time.Minute,
		MinInterval:       6 * time.Second, // ~10 requests per minute
		BackoffBase:       30 * time.Second,
		MaxBackoff:        30 * time.Minute,
	}

	return limiter
}

// WaitForPermission blocks until it's safe to make a request to the specified provider
func (rl *ProviderRateLimiter) WaitForPermission(ctx context.Context, provider string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	config, exists := rl.rateLimits[provider]
	if !exists {
		// Unknown provider, allow with minimal throttling
		config = RateLimitConfig{
			RequestsPerWindow: 60,
			WindowSize:        time.Minute,
			MinInterval:       1 * time.Second,
			BackoffBase:       5 * time.Second,
			MaxBackoff:        2 * time.Minute,
		}
		rl.rateLimits[provider] = config
	}

	state, exists := rl.providers[provider]
	if !exists {
		state = &ProviderLimitState{
			windowStart: time.Now(),
		}
		rl.providers[provider] = state
	}

	now := time.Now()

	// Check if we need to reset the window
	if now.Sub(state.windowStart) >= config.WindowSize {
		state.windowStart = now
		state.requestCount = 0
	}

	// Calculate wait time based on multiple factors
	waitTime := rl.calculateWaitTime(state, config, now)

	if waitTime > 0 {
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			rl.mu.Lock() // Re-acquire lock before returning
			return ctx.Err()
		case <-time.After(waitTime):
			// Wait completed successfully
		}

		rl.mu.Lock() // Re-acquire lock
	}

	state.lastRequest = time.Now()
	state.requestCount++

	return nil
}

// calculateWaitTime determines how long to wait before allowing the next request
func (rl *ProviderRateLimiter) calculateWaitTime(state *ProviderLimitState, config RateLimitConfig, now time.Time) time.Duration {
	var waitTimes []time.Duration

	// 1. Minimum interval between requests
	timeSinceLastRequest := now.Sub(state.lastRequest)
	if timeSinceLastRequest < config.MinInterval {
		waitTimes = append(waitTimes, config.MinInterval-timeSinceLastRequest)
	}

	// 2. Rate limit window constraint
	if state.requestCount >= config.RequestsPerWindow {
		timeUntilWindowReset := config.WindowSize - now.Sub(state.windowStart)
		if timeUntilWindowReset > 0 {
			waitTimes = append(waitTimes, timeUntilWindowReset)
		}
	}

	// 3. Exponential backoff for consecutive failures
	if state.currentBackoff > 0 {
		waitTimes = append(waitTimes, state.currentBackoff)
	}

	// Return the maximum wait time required
	var maxWait time.Duration
	for _, wait := range waitTimes {
		if wait > maxWait {
			maxWait = wait
		}
	}

	return maxWait
}

// RecordSuccess records a successful API call
func (rl *ProviderRateLimiter) RecordSuccess(provider string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if state, exists := rl.providers[provider]; exists {
		state.consecutiveFails = 0
		state.currentBackoff = 0
	}
}

// RecordFailure records a failed API call and applies backoff
func (rl *ProviderRateLimiter) RecordFailure(provider string, isRateLimit bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	config, exists := rl.rateLimits[provider]
	if !exists {
		return
	}

	state, exists := rl.providers[provider]
	if !exists {
		state = &ProviderLimitState{}
		rl.providers[provider] = state
	}

	state.consecutiveFails++

	if isRateLimit {
		// Apply aggressive backoff for rate limit errors
		state.currentBackoff = config.BackoffBase * time.Duration(1<<min(state.consecutiveFails, 4))
	} else {
		// Apply gentler backoff for other errors
		state.currentBackoff = config.BackoffBase * time.Duration(state.consecutiveFails)
	}

	// Cap the backoff at maximum
	if state.currentBackoff > config.MaxBackoff {
		state.currentBackoff = config.MaxBackoff
	}
}

func (rl *ProviderRateLimiter) GetProviderStatus(provider string) ProviderStatus {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	config, configExists := rl.rateLimits[provider]
	state, stateExists := rl.providers[provider]

	if !configExists || !stateExists {
		return ProviderStatus{
			Provider:         provider,
			Available:        true,
			RequestsInWindow: 0,
		}
	}

	now := time.Now()
	windowTimeRemaining := config.WindowSize - now.Sub(state.windowStart)
	if windowTimeRemaining < 0 {
		windowTimeRemaining = 0
	}

	nextRequestTime := state.lastRequest.Add(config.MinInterval)
	if state.currentBackoff > 0 {
		backoffEndTime := state.lastRequest.Add(state.currentBackoff)
		if backoffEndTime.After(nextRequestTime) {
			nextRequestTime = backoffEndTime
		}
	}

	var nextRequestDelay time.Duration
	if nextRequestTime.After(now) {
		nextRequestDelay = nextRequestTime.Sub(now)
	}

	return ProviderStatus{
		Provider:            provider,
		Available:           nextRequestDelay == 0 && state.requestCount < config.RequestsPerWindow,
		RequestsInWindow:    state.requestCount,
		MaxRequestsInWindow: config.RequestsPerWindow,
		WindowTimeRemaining: windowTimeRemaining,
		NextRequestDelay:    nextRequestDelay,
		ConsecutiveFailures: state.consecutiveFails,
		CurrentBackoff:      state.currentBackoff,
	}
}

// ProviderStatus represents the current rate limiting status for a provider
type ProviderStatus struct {
	Provider            string
	Available           bool
	RequestsInWindow    int
	MaxRequestsInWindow int
	WindowTimeRemaining time.Duration
	NextRequestDelay    time.Duration
	ConsecutiveFailures int
	CurrentBackoff      time.Duration
}

// String returns a human-readable status description
func (ps ProviderStatus) String() string {
	if ps.Available {
		return fmt.Sprintf("%s: Available (%d/%d requests used)",
			ps.Provider, ps.RequestsInWindow, ps.MaxRequestsInWindow)
	}

	reasons := make([]string, 0)
	if ps.RequestsInWindow >= ps.MaxRequestsInWindow {
		reasons = append(reasons, fmt.Sprintf("rate limited (%d/%d, reset in %v)",
			ps.RequestsInWindow, ps.MaxRequestsInWindow, ps.WindowTimeRemaining))
	}
	if ps.NextRequestDelay > 0 {
		reasons = append(reasons, fmt.Sprintf("throttled (wait %v)", ps.NextRequestDelay))
	}
	if ps.CurrentBackoff > 0 {
		reasons = append(reasons, fmt.Sprintf("backoff (%v, %d failures)",
			ps.CurrentBackoff, ps.ConsecutiveFailures))
	}

	reasonsStr := "unknown"
	if len(reasons) > 0 {
		reasonsStr = reasons[0] // Show primary reason
	}

	return fmt.Sprintf("%s: Unavailable (%s)", ps.Provider, reasonsStr)
}

func (rl *ProviderRateLimiter) UpdateRateLimit(provider string, config RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.rateLimits[provider] = config
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
