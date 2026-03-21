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

// GetNextAdminToken finds an active token for the given platform with remaining rate limit.
// It uses round-robin logic by selecting the one least recently updated.
// The SELECT + UPDATE is wrapped in a transaction with FOR UPDATE to prevent race conditions.
func GetNextAdminToken(ctx context.Context, platform string) (*AdminForgeToken, error) {
	var result *AdminForgeToken
	err := db.WithTx(ctx, func(ctx context.Context) error {
		token := new(AdminForgeToken)
		has, err := db.GetEngine(ctx).
			Where("platform=?", platform).
			And("is_active=?", true).
			And("rate_limit_remaining > ?", 0).
			Asc("updated_unix").
			ForUpdate().
			Get(token)
		if err != nil {
			return err
		}
		if !has {
			return nil // No active tokens available
		}

		// Update the token to push it to the back of the queue (Round-Robin)
		token.RequestCount++
		_, err = db.GetEngine(ctx).ID(token.ID).Cols("request_count", "updated_unix").Update(token)
		if err != nil {
			return err
		}
		result = token
		return nil
	})
	return result, err
}

// IncrementUserQuota increments the daily quota for a user.
// Creates a new record for today if it doesn't exist.
// Wrapped in a transaction with FOR UPDATE to prevent race conditions
// where concurrent requests both see !has and both INSERT.
func IncrementUserQuota(ctx context.Context, userID int64) error {
	today := time.Now().Format("2006-01-02")
	return db.WithTx(ctx, func(ctx context.Context) error {
		quota := new(UserTokenQuota)
		has, err := db.GetEngine(ctx).Where("user_id=? AND date_str=?", userID, today).ForUpdate().Get(quota)
		if err != nil {
			return err
		}

		if !has {
			// First request today, create new quota record
			quota = &UserTokenQuota{
				UserID:        userID,
				DateStr:       today,
				RequestsUsed:  1,
				RequestsLimit: 50, // Default limit
			}
			_, err = db.GetEngine(ctx).Insert(quota)
			return err
		}

		// Increment existing quota
		quota.RequestsUsed++
		_, err = db.GetEngine(ctx).ID(quota.ID).Cols("requests_used").Update(quota)
		return err
	})
}

// GetOriginalAuthorsFromRepository extracts distinct OriginalAuthorID and OriginalAuthor
// stored in the local database (Issues and Comments) after a repository migration.
// This is used for mapping user accounts between Forge platforms.
func GetOriginalAuthorsFromRepository(ctx context.Context, repoID int64) ([]*ForgeUserMapping, error) {
	type authorResult struct {
		OriginalAuthorID int64
		OriginalAuthor   string
	}

	// 1. Get from Issues
	var issueAuthors []authorResult
	err := db.GetEngine(ctx).Table("issue").
		Cols("original_author_id", "original_author").
		Where("repo_id = ? AND original_author_id > 0", repoID).
		GroupBy("original_author_id, original_author").
		Find(&issueAuthors)
	if err != nil {
		return nil, err
	}

	// 2. Get from Comments
	var commentAuthors []authorResult
	err = db.GetEngine(ctx).Table("comment").
		Cols("original_author_id", "original_author").
		Where("issue_id IN (SELECT id FROM issue WHERE repo_id = ?) AND original_author_id > 0", repoID).
		GroupBy("original_author_id, original_author").
		Find(&commentAuthors)
	if err != nil {
		return nil, err
	}

	// 3. Get from Reviews
	var reviewAuthors []authorResult
	err = db.GetEngine(ctx).Table("review").
		Cols("original_author_id", "original_author").
		Where("issue_id IN (SELECT id FROM issue WHERE repo_id = ?) AND original_author_id > 0", repoID).
		GroupBy("original_author_id, original_author").
		Find(&reviewAuthors)
	if err != nil && !db.IsErrNotExist(err) {
		return nil, err
	}

	// Deduplicate in memory
	seen := make(map[int64]bool)
	var mappings []*ForgeUserMapping

	allAuthors := append(issueAuthors, commentAuthors...)
	allAuthors = append(allAuthors, reviewAuthors...)

	for _, a := range allAuthors {
		if !seen[a.OriginalAuthorID] {
			seen[a.OriginalAuthorID] = true
			mappings = append(mappings, &ForgeUserMapping{
				ForgeUserID: a.OriginalAuthorID,
				ForgeLogin:  a.OriginalAuthor,
				GiteaUserID: 0, // Will be mapped later in Phase 10
			})
		}
	}

	return mappings, nil
}

// === END CUSTOM: forge-bridge ===
