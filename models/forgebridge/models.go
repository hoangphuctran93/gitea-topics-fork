// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

// AdminForgeToken represents a forge token (GitHub, Gitea, etc.) from the shared admin pool.
type AdminForgeToken struct {
	ID                 int64              `xorm:"pk autoincr"`
	Platform           string             `xorm:"VARCHAR(20) NOT NULL INDEX"` // 'github', 'gitea'
	Label              string             `xorm:"VARCHAR(100)"`
	TokenEncrypted     string             `xorm:"TEXT NOT NULL"`
	RequestCount       int64              `xorm:"DEFAULT 0"`
	RateLimitRemaining int                `xorm:"DEFAULT 5000"`
	IsActive           bool               `xorm:"DEFAULT true"`
	CreatedUnix        timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix        timeutil.TimeStamp `xorm:"updated"`
}

// UserForgeToken represents a personal forge token provided by a user.
type UserForgeToken struct {
	ID                int64  `xorm:"pk autoincr"`
	UserID            int64  `xorm:"UNIQUE(user_platform) NOT NULL"`
	Platform          string `xorm:"VARCHAR(20) UNIQUE(user_platform) NOT NULL"` // 'github', 'gitea'
	TokenEncrypted    string `xorm:"TEXT NOT NULL"`
	ForgeUsername     string `xorm:"VARCHAR(255)"`
	ShareForSearch    bool   `xorm:"DEFAULT false"`
	ShareForCloneInfo bool   `xorm:"DEFAULT false"`
	Source            string `xorm:"VARCHAR(20) DEFAULT 'manual'"`
	LastUsedUnix      timeutil.TimeStamp
	CreatedUnix       timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix       timeutil.TimeStamp `xorm:"updated"`
}

// ForgeUserMapping maps a Gitea user to an external forge user to prevent duplicates.
type ForgeUserMapping struct {
	ID             int64              `xorm:"pk autoincr"`
	Platform       string             `xorm:"VARCHAR(20) NOT NULL INDEX"`
	GiteaUserID    int64              `xorm:"INDEX"` // Can be 0 if not yet mapped
	ForgeUserID    int64              `xorm:"INDEX NOT NULL"`
	ForgeLogin     string             `xorm:"VARCHAR(255) NOT NULL"`
	LastSyncedUnix timeutil.TimeStamp `xorm:"DEFAULT 0"`
	CreatedUnix    timeutil.TimeStamp `xorm:"created"`
}

// UserTokenQuota tracks daily quota usage for users without their own tokens.
type UserTokenQuota struct {
	ID            int64  `xorm:"pk autoincr"`
	UserID        int64  `xorm:"UNIQUE(user_date) NOT NULL"`
	DateStr       string `xorm:"VARCHAR(10) UNIQUE(user_date) NOT NULL"` // Format: YYYY-MM-DD
	RequestsUsed  int    `xorm:"DEFAULT 0"`
	RequestsLimit int    `xorm:"DEFAULT 50"`
}

func init() {
	db.RegisterModel(new(AdminForgeToken))
	db.RegisterModel(new(UserForgeToken))
	db.RegisterModel(new(ForgeUserMapping))
	db.RegisterModel(new(UserTokenQuota))
}

// === END CUSTOM: forge-bridge ===
