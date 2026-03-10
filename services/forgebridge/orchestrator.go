// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"context"
	"errors"
	"fmt"

	forgebridge_model "code.gitea.io/gitea/models/forgebridge"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

var (
	ErrQuotaExceeded = errors.New("user has exceeded daily request quota")
	ErrNoTokensAvailable = errors.New("no active tokens available")
)

// GetTokenForUser is the main orchestrator function.
// It tries to find a user's personal token first.
// If not found, it checks the daily quota and then borrows a round-robin admin token.
func GetTokenForUser(ctx context.Context, userID int64, platform string) (token string, source string, err error) {
	// 1. Check for personal user token
	userToken := &forgebridge_model.UserForgeToken{}
	hasUserToken, err := forgebridge_model.GetUserForgeToken(ctx, userID, platform, userToken)
	if err != nil {
		return "", "", fmt.Errorf("checking user token: %w", err)
	}

	if hasUserToken {
		// Decrypt and return user's own token
		decrypted, err := forgebridge_model.DecryptToken(userToken.TokenEncrypted)
		if err != nil {
			return "", "", fmt.Errorf("decrypting user token: %w", err)
		}
		// Update last used time
		// (Assuming we want to track this, we'd need an update function for LastUsedUnix)
		return decrypted, "user", nil
	}

	// 2. Check Daily Quota for borrowing
	quota := &forgebridge_model.UserTokenQuota{}
	hasQuotaRecord, err := forgebridge_model.GetTodayUserQuota(ctx, userID, quota)
	if err != nil {
		return "", "", fmt.Errorf("checking daily quota: %w", err)
	}

	if hasQuotaRecord && quota.RequestsUsed >= quota.RequestsLimit {
		return "", "", ErrQuotaExceeded
	}

	// 3. Get next available Admin token via Round-Robin
	adminToken, err := forgebridge_model.GetNextAdminToken(ctx, platform)
	if err != nil {
		return "", "", fmt.Errorf("getting admin token: %w", err)
	}
	if adminToken == nil {
		return "", "", ErrNoTokensAvailable
	}

	// 4. Decrypt Admin token
	decryptedAdminToken, err := forgebridge_model.DecryptToken(adminToken.TokenEncrypted)
	if err != nil {
		return "", "", fmt.Errorf("decrypting admin token: %w", err)
	}

	// 5. Increment user's daily quota usage
	err = forgebridge_model.IncrementUserQuota(ctx, userID)
	if err != nil {
		// Just log it in real world, but returning error for strictness
		return "", "", fmt.Errorf("incrementing quota: %w", err)
	}

	return decryptedAdminToken, "admin", nil
}

// === END CUSTOM: forge-bridge ===
