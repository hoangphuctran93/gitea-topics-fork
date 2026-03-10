// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"net/http"
	"strconv"
	"time"
)

// RateLimitInfo stores rate limit data parsed from GitHub headers
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetUnix int64
}

// ParseRateLimitHeaders extracts rate limit info from standard GitHub response headers
func ParseRateLimitHeaders(header http.Header) *RateLimitInfo {
	limitStr := header.Get("X-RateLimit-Limit")
	remainingStr := header.Get("X-RateLimit-Remaining")
	resetStr := header.Get("X-RateLimit-Reset")

	if limitStr == "" || remainingStr == "" || resetStr == "" {
		return nil
	}

	limit, _ := strconv.Atoi(limitStr)
	remaining, _ := strconv.Atoi(remainingStr)
	resetUnix, _ := strconv.ParseInt(resetStr, 10, 64)

	return &RateLimitInfo{
		Limit:     limit,
		Remaining: remaining,
		ResetUnix: resetUnix,
	}
}

// CalculateAdaptiveDelay returns an advised sleep duration based on remaining quota
func CalculateAdaptiveDelay(info *RateLimitInfo, baseDelay time.Duration) time.Duration {
	if info == nil {
		return baseDelay
	}

	// If we are out of quota, we should wait until reset
	if info.Remaining <= 0 {
		now := time.Now().Unix()
		waitSecs := info.ResetUnix - now
		if waitSecs > 0 {
			// Cap at 1 hour max
			if waitSecs > 3600 {
				waitSecs = 3600
			}
			return time.Duration(waitSecs) * time.Second
		}
		return baseDelay * 2 // Fallback
	}

	// Adaptive backoff: if remaining is low, sleep longer
	if info.Remaining < 50 {
		return baseDelay * 5
	} else if info.Remaining < 200 {
		return baseDelay * 2
	} else if info.Remaining > 1000 {
		return baseDelay / 2 // Speed up if plenty
	}

	return baseDelay
}

// === END CUSTOM: forge-bridge ===
