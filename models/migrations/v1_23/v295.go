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

	type GithubUserMapping struct {
		ID             int64              `xorm:"pk autoincr"`
		GiteaUserID    int64              `xorm:"INDEX"`
		GithubUserID   int64              `xorm:"UNIQUE NOT NULL"`
		GithubLogin    string             `xorm:"VARCHAR(255) NOT NULL"`
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
		new(AdminGithubToken),
		new(UserGithubToken),
		new(GithubUserMapping),
		new(UserTokenQuota),
	)
	// === END CUSTOM: forge-bridge ===
}
