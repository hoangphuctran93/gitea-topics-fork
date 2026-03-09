// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package forgebridge

import (
	"time"

	"code.gitea.io/gitea/models/db"
)

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

func GetUserGithubToken(userID int64, token *UserGithubToken) (bool, error) {
	return db.GetEngine(db.DefaultContext).Where("user_id=?", userID).Get(token)
}

func InsertUserGithubToken(token *UserGithubToken) error {
	_, err := db.GetEngine(db.DefaultContext).Insert(token)
	return err
}

func UpdateUserGithubToken(token *UserGithubToken) error {
	_, err := db.GetEngine(db.DefaultContext).ID(token.ID).AllCols().Update(token)
	return err
}

func DeleteUserGithubToken(userID int64) error {
	_, err := db.GetEngine(db.DefaultContext).Where("user_id=?", userID).Delete(new(UserGithubToken))
	return err
}

func GetTodayUserQuota(userID int64, quota *UserTokenQuota) (bool, error) {
	today := time.Now().Format("2006-01-02")
	return db.GetEngine(db.DefaultContext).Where("user_id=? AND date_str=?", userID, today).Get(quota)
}

func GetAllAdminTokens() ([]*AdminGithubToken, error) {
	tokens := make([]*AdminGithubToken, 0)
	err := db.GetEngine(db.DefaultContext).Find(&tokens)
	return tokens, err
}

func InsertAdminGithubToken(token *AdminGithubToken) error {
	_, err := db.GetEngine(db.DefaultContext).Insert(token)
	return err
}

func DeleteAdminGithubToken(id int64) error {
	_, err := db.GetEngine(db.DefaultContext).ID(id).Delete(new(AdminGithubToken))
	return err
}

// === END CUSTOM: forge-bridge ===
