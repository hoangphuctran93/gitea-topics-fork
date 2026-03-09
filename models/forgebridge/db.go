// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"context"
	"time"

	"code.gitea.io/gitea/models/db"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

func GetUserForgeToken(ctx context.Context, userID int64, platform string, token *UserForgeToken) (bool, error) {
	return db.GetEngine(ctx).Where("user_id=? AND platform=?", userID, platform).Get(token)
}

func InsertUserForgeToken(ctx context.Context, token *UserForgeToken) error {
	_, err := db.GetEngine(ctx).Insert(token)
	return err
}

func UpdateUserForgeToken(ctx context.Context, token *UserForgeToken) error {
	_, err := db.GetEngine(ctx).ID(token.ID).AllCols().Update(token)
	return err
}

func DeleteUserForgeToken(ctx context.Context, userID int64, platform string) error {
	_, err := db.GetEngine(ctx).Where("user_id=? AND platform=?", userID, platform).Delete(new(UserForgeToken))
	return err
}

func GetTodayUserQuota(ctx context.Context, userID int64, quota *UserTokenQuota) (bool, error) {
	today := time.Now().Format("2006-01-02")
	return db.GetEngine(ctx).Where("user_id=? AND date_str=?", userID, today).Get(quota)
}

func GetAllAdminTokens(ctx context.Context, platform string) ([]*AdminForgeToken, error) {
	tokens := make([]*AdminForgeToken, 0)
	engine := db.GetEngine(ctx)
	if platform != "" {
		engine = engine.Where("platform=?", platform)
	}
	err := engine.Find(&tokens)
	return tokens, err
}

func InsertAdminForgeToken(ctx context.Context, token *AdminForgeToken) error {
	_, err := db.GetEngine(ctx).Insert(token)
	return err
}

func DeleteAdminForgeToken(ctx context.Context, id int64) error {
	_, err := db.GetEngine(ctx).ID(id).Delete(new(AdminForgeToken))
	return err
}

// === END CUSTOM: forge-bridge ===
