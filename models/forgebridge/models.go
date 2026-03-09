// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

// AdminGithubToken represents a GitHub token from the shared admin pool.
type AdminGithubToken struct {
	ID                 int64              `xorm:"pk autoincr"`
	Label              string             `xorm:"VARCHAR(100)"`
	TokenEncrypted     string             `xorm:"TEXT NOT NULL"`
	RequestCount       int64              `xorm:"DEFAULT 0"`
	RateLimitRemaining int                `xorm:"DEFAULT 5000"`
	IsActive           bool               `xorm:"DEFAULT true"`
	CreatedUnix        timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix        timeutil.TimeStamp `xorm:"updated"`
}

// UserGithubToken represents a personal GitHub token provided by a user.
type UserGithubToken struct {
	ID                int64  `xorm:"pk autoincr"`
	UserID            int64  `xorm:"UNIQUE NOT NULL"`
	TokenEncrypted    string `xorm:"TEXT NOT NULL"`
	GithubUsername    string `xorm:"VARCHAR(255)"`
	ShareForSearch    bool   `xorm:"DEFAULT false"`
	ShareForCloneInfo bool   `xorm:"DEFAULT false"`
	Source            string `xorm:"VARCHAR(20) DEFAULT 'manual'"`
	LastUsedUnix      timeutil.TimeStamp
	CreatedUnix       timeutil.TimeStamp `xorm:"created"`
	UpdatedUnix       timeutil.TimeStamp `xorm:"updated"`
}

// GithubUserMapping maps a Gitea user to a GitHub user to prevent duplicates.
type GithubUserMapping struct {
	ID             int64              `xorm:"pk autoincr"`
	GiteaUserID    int64              `xorm:"INDEX"` // Can be 0 if not yet mapped
	GithubUserID   int64              `xorm:"UNIQUE NOT NULL"`
	GithubLogin    string             `xorm:"VARCHAR(255) NOT NULL"`
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
	db.RegisterModel(new(AdminGithubToken))
	db.RegisterModel(new(UserGithubToken))
	db.RegisterModel(new(GithubUserMapping))
	db.RegisterModel(new(UserTokenQuota))
}

// === END CUSTOM: forge-bridge ===
