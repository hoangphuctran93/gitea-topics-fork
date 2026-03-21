// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_23 // Assuming the next migration batch is for v1.23, adjusting based on current Gitea v1.22.x fork

import (
	"code.gitea.io/gitea/modules/timeutil"

	"xorm.io/xorm"
)

func AddForgeBridgeTables(x *xorm.Engine) error {
	// === BEGIN CUSTOM: forge-bridge ===
	// Upstream-safe: false | Author: hoangphuctran93

	type AdminForgeToken struct {
		ID                   int64  `xorm:"pk autoincr"`
		TokenName            string `xorm:"UNIQUE NOT NULL"` // Name for UI display
		Platform             string `xorm:"INDEX NOT NULL"`  // "github", "gitlab", etc.
		TokenVal             string `xorm:"TEXT NOT NULL"`   // The actual token or encrypted string
		RateLimitTotal       int    `xorm:"NOT NULL DEFAULT 5000"`
		RateLimitRemaining   int    `xorm:"NOT NULL DEFAULT 5000"`
		RateLimitResetUnix   int64  // Unix timestamp when reset occurs
		IsActive             bool   `xorm:"NOT NULL DEFAULT true"` // Toggle to enable/disable
		EncryptionSalt       string // Optional salt if encrypted
		EncryptionMethod     string // Method used for encryption (e.g., "aes-256-gcm")
		LastRateLimitCheck   int64  // When the rate limit was last verified
		ConsecutiveFailures  int    // Counter for automated disabling on failures
		CreatedUnix          int64  `xorm:"created"`
		UpdatedUnix          int64  `xorm:"updated"`
		RequestCount         int    `xorm:"NOT NULL DEFAULT 0"`    // Analytics: Total requests made with this token
		SuccessCount         int    `xorm:"NOT NULL DEFAULT 0"`    // Analytics: Successful requests
		TotalRateLimitTokens int    `xorm:"NOT NULL DEFAULT 5000"` // Store the total allocation size observed from API
	}

	type UserForgeToken struct {
		ID                int64  `xorm:"pk autoincr"`
		UserID            int64  `xorm:"UNIQUE NOT NULL"`
		Platform          string `xorm:"INDEX NOT NULL"` // "github", "gitlab", etc.
		TokenEncrypted    string `xorm:"TEXT NOT NULL"`
		ForgeUsername     string `xorm:"VARCHAR(255)"`
		ShareForSearch    bool   `xorm:"DEFAULT false"`
		ShareForCloneInfo bool   `xorm:"DEFAULT false"`
		Source            string `xorm:"VARCHAR(20) DEFAULT 'manual'"`
		LastUsedUnix      timeutil.TimeStamp
		CreatedUnix       timeutil.TimeStamp `xorm:"created"`
		UpdatedUnix       timeutil.TimeStamp `xorm:"updated"`
	}

	type ForgeUserMapping struct {
		ID             int64              `xorm:"pk autoincr"`
		Platform       string             `xorm:"VARCHAR(20) NOT NULL INDEX"`
		GiteaUserID    int64              `xorm:"INDEX"` // Can be 0 if not yet mapped
		ForgeUserID    int64              `xorm:"INDEX NOT NULL"`
		ForgeLogin     string             `xorm:"VARCHAR(255) NOT NULL"`
		LastSyncedUnix timeutil.TimeStamp `xorm:"DEFAULT 0"`
		CreatedUnix    timeutil.TimeStamp `xorm:"created"`
	}

	type UserTokenQuota struct {
		ID            int64  `xorm:"pk autoincr"`
		UserID        int64  `xorm:"UNIQUE(user_date) NOT NULL"`
		DateStr       string `xorm:"VARCHAR(10) UNIQUE(user_date) NOT NULL"`
		RequestsUsed  int    `xorm:"DEFAULT 0"`
		RequestsLimit int    `xorm:"DEFAULT 50"`
	}

	return x.Sync(
		new(AdminForgeToken),
		new(UserForgeToken),
		new(ForgeUserMapping),
		new(UserTokenQuota),
	)
	// === END CUSTOM: forge-bridge ===
}
