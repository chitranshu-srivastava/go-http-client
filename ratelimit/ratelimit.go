package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter wraps golang.org/x/time/rate.Limiter with additional functionality
type RateLimiter struct {
	limiter *rate.Limiter
	enabled bool
	mu      sync.RWMutex
}

// Config holds rate limiting configuration
type Config struct {
	Rate    string // Rate string like "10/s" or "100/30s"
	Enabled bool   // Whether rate limiting is enabled
}

// New creates a new RateLimiter from a rate string
func New(rateStr string) (*RateLimiter, error) {
	if rateStr == "" {
		return &RateLimiter{enabled: false}, nil
	}

	limit, burst, err := parseRate(rateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid rate format: %w", err)
	}

	limiter := rate.NewLimiter(limit, burst)
	return &RateLimiter{
		limiter: limiter,
		enabled: true,
	}, nil
}

// parseRate parses rate strings like "10/s", "100/30s", "50/m", "1000/h"
func parseRate(rateStr string) (rate.Limit, int, error) {
	parts := strings.Split(rateStr, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("rate must be in format 'requests/duration' (e.g., '10/s', '100/30s')")
	}

	requests, err := strconv.Atoi(parts[0])
	if err != nil || requests <= 0 {
		return 0, 0, fmt.Errorf("requests must be a positive integer")
	}

	duration, err := parseDuration(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid duration: %w", err)
	}

	// Calculate rate per second
	limit := rate.Limit(float64(requests) / duration.Seconds())
	
	// Set burst to requests count, allowing short bursts up to the limit
	burst := requests

	return limit, burst, nil
}

// parseDuration handles duration strings like "s", "30s", "m", "h"
func parseDuration(durStr string) (time.Duration, error) {
	// Handle simple cases first
	switch durStr {
	case "s":
		return time.Second, nil
	case "m":
		return time.Minute, nil
	case "h":
		return time.Hour, nil
	}

	// Try parsing as standard duration
	duration, err := time.ParseDuration(durStr)
	if err != nil {
		return 0, fmt.Errorf("duration must be a valid time duration (e.g., 's', '30s', 'm', 'h') or a number followed by a unit")
	}

	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}

	return duration, nil
}

// Allow waits for permission to proceed with the request
// Returns nil if the request is allowed, or an error if rate limited
func (rl *RateLimiter) Allow() error {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if !rl.enabled {
		return nil
	}

	if !rl.limiter.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// Wait blocks until the request can proceed or context is cancelled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if !rl.enabled {
		return nil
	}

	return rl.limiter.Wait(ctx)
}

// SetRate updates the rate limit
func (rl *RateLimiter) SetRate(rateStr string) error {
	if rateStr == "" {
		rl.mu.Lock()
		rl.enabled = false
		rl.mu.Unlock()
		return nil
	}

	limit, burst, err := parseRate(rateStr)
	if err != nil {
		return fmt.Errorf("invalid rate format: %w", err)
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.limiter == nil {
		rl.limiter = rate.NewLimiter(limit, burst)
	} else {
		rl.limiter.SetLimit(limit)
		rl.limiter.SetBurst(burst)
	}
	rl.enabled = true

	return nil
}

// IsEnabled returns whether rate limiting is enabled
func (rl *RateLimiter) IsEnabled() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.enabled
}

// Stats returns rate limiting statistics
func (rl *RateLimiter) Stats() map[string]any {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := make(map[string]any)
	stats["enabled"] = rl.enabled

	if rl.enabled && rl.limiter != nil {
		stats["limit"] = float64(rl.limiter.Limit())
		stats["burst"] = rl.limiter.Burst()
		stats["tokens"] = rl.limiter.Tokens()
	}

	return stats
}