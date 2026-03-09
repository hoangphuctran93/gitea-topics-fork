// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"time"

	"code.gitea.io/gitea/models/db"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

func GetUserForgeToken(userID int64, platform string, token *UserForgeToken) (bool, error) {
	return db.GetEngine(db.DefaultContext).Where("user_id=? AND platform=?", userID, platform).Get(token)
}

func InsertUserForgeToken(token *UserForgeToken) error {
	_, err := db.GetEngine(db.DefaultContext).Insert(token)
	return err
}

func UpdateUserForgeToken(token *UserForgeToken) error {
	_, err := db.GetEngine(db.DefaultContext).ID(token.ID).AllCols().Update(token)
	return err
}

func DeleteUserForgeToken(userID int64, platform string) error {
	_, err := db.GetEngine(db.DefaultContext).Where("user_id=? AND platform=?", userID, platform).Delete(new(UserForgeToken))
	return err
}

func GetTodayUserQuota(userID int64, quota *UserTokenQuota) (bool, error) {
	today := time.Now().Format("2006-01-02")
	return db.GetEngine(db.DefaultContext).Where("user_id=? AND date_str=?", userID, today).Get(quota)
}

func GetAllAdminTokens(platform string) ([]*AdminForgeToken, error) {
	tokens := make([]*AdminForgeToken, 0)
	engine := db.GetEngine(db.DefaultContext)
	if platform != "" {
		engine = engine.Where("platform=?", platform)
	}
	err := engine.Find(&tokens)
	return tokens, err
}

func InsertAdminForgeToken(token *AdminForgeToken) error {
	_, err := db.GetEngine(db.DefaultContext).Insert(token)
	return err
}

func DeleteAdminForgeToken(id int64) error {
	_, err := db.GetEngine(db.DefaultContext).ID(id).Delete(new(AdminForgeToken))
	return err
}

// === END CUSTOM: forge-bridge ===
