package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		name        string
		rateStr     string
		expectError bool
	}{
		{"Valid rate 10/s", "10/s", false},
		{"Valid rate 100/30s", "100/30s", false},
		{"Valid rate 50/m", "50/m", false},
		{"Valid rate 1000/h", "1000/h", false},
		{"Invalid format", "invalid", true},
		{"Empty string", "", false}, // Should create disabled limiter
		{"Zero requests", "0/s", true},
		{"Negative requests", "-10/s", true},
		{"Invalid duration", "10/xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter, err := New(tt.rateStr)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for rate string: %s", tt.rateStr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for rate string %s: %v", tt.rateStr, err)
			}
			if tt.rateStr == "" && limiter != nil && limiter.IsEnabled() {
				t.Error("Empty rate string should create disabled limiter")
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	// Test rate limiting with 2/s (2 requests per second)
	limiter, err := New("2/s")
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	if !limiter.IsEnabled() {
		t.Error("Rate limiter should be enabled")
	}

	// First two requests should pass immediately
	if err := limiter.Allow(); err != nil {
		t.Errorf("First request should be allowed: %v", err)
	}
	
	if err := limiter.Allow(); err != nil {
		t.Errorf("Second request should be allowed: %v", err)
	}

	// Third request should be rate limited
	if err := limiter.Allow(); err == nil {
		t.Error("Third request should be rate limited")
	}
}

func TestSetRate(t *testing.T) {
	limiter, err := New("")
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	if limiter.IsEnabled() {
		t.Error("Initially disabled limiter should not be enabled")
	}

	// Set a rate
	err = limiter.SetRate("5/s")
	if err != nil {
		t.Errorf("Failed to set rate: %v", err)
	}

	if !limiter.IsEnabled() {
		t.Error("Rate limiter should be enabled after setting rate")
	}

	// Disable again
	err = limiter.SetRate("")
	if err != nil {
		t.Errorf("Failed to disable rate limiter: %v", err)
	}

	if limiter.IsEnabled() {
		t.Error("Rate limiter should be disabled after setting empty rate")
	}
}

func TestWait(t *testing.T) {
	limiter, err := New("1/s")
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	ctx := context.Background()
	
	// First request should not wait
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Errorf("First wait should not error: %v", err)
	}
	elapsed := time.Since(start)
	
	if elapsed > 10*time.Millisecond {
		t.Errorf("First request should not wait, but took %v", elapsed)
	}
}

func TestStats(t *testing.T) {
	limiter, err := New("10/s")
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	stats := limiter.Stats()
	
	if !stats["enabled"].(bool) {
		t.Error("Stats should show limiter as enabled")
	}

	if stats["limit"].(float64) != 10.0 {
		t.Errorf("Expected limit of 10.0, got %v", stats["limit"])
	}

	if stats["burst"].(int) != 10 {
		t.Errorf("Expected burst of 10, got %v", stats["burst"])
	}
}