// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/forgebridge"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/timeutil"
)

// GitHubUserInfo represents the JSON response from api.github.com/users/:username
type GitHubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	Blog      string `json:"blog"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
}

func syncGitHubProfile(ctx context.Context, task *SyncTask) error {
	// 1. Extract GitHub Username from OriginalURL
	// e.g. https://github.com/octocat/Hello-World.git -> octocat
	parsedURL, err := url.Parse(task.OriginalURL)
	if err != nil {
		return fmt.Errorf("failed to parse original URL: %v", err)
	}

	// Basic check: only proceed if the host is github.com
	if !strings.EqualFold(parsedURL.Host, "github.com") {
		log.Trace("ForgeBridge: OriginalURL %s is not github.com, skipping sync", task.OriginalURL)
		return nil
	}

	pathParts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/"), "/")
	if len(pathParts) < 1 {
		return fmt.Errorf("invalid github URL path: %s", parsedURL.Path)
	}
	githubUsername := pathParts[0]
	if githubUsername == "" {
		return fmt.Errorf("empty github username from URL: %s", task.OriginalURL)
	}

	// 2. Resolve Token
	var tokenStr string
	var tokenSource string
	tokenStr, tokenSource, err = GetTokenForUser(ctx, task.RepoOwnerID, "github")
	if err != nil && err != ErrNoTokensAvailable {
		log.Warn("ForgeBridge: GetTokenForUser error for RepoOwnerID %d: %v", task.RepoOwnerID, err)
	}
	log.Trace("ForgeBridge: Using token from %s for github sync%s", tokenSource, githubUsername)

	// 3. Fetch GitHub Profile
	reqTarget := fmt.Sprintf("https://api.github.com/users/%s", githubUsername)
	req, err := http.NewRequestWithContext(ctx, "GET", reqTarget, nil)
	if err != nil {
		return err
	}
	if tokenStr != "" {
		req.Header.Set("Authorization", "Bearer "+tokenStr)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("github api request failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse Rate Limit
	rateLimit := ParseRateLimitHeaders(resp.Header)
	delay := CalculateAdaptiveDelay(rateLimit, 1*time.Second)
	if delay > 0 {
		log.Trace("ForgeBridge: Rate limit aware delay: %v (remaining: %d)", delay, rateLimit.Remaining)
		time.Sleep(delay)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api returned non-200 status: %d for user %s", resp.StatusCode, githubUsername)
	}

	var ghUser GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return fmt.Errorf("failed to decode github response: %v", err)
	}

	// 4. Update Database
	return db.WithTx(ctx, func(ctx context.Context) error {
		// Check Mapping Deduplication
		exists, err := db.GetEngine(ctx).Where("platform = ? AND forge_user_id = ?", "github", ghUser.ID).Get(new(forgebridge.ForgeUserMapping))
		if err != nil {
			return err
		}
		if exists {
			// Check if last sync was < 24h
			var mapping forgebridge.ForgeUserMapping
			if _, err := db.GetEngine(ctx).Where("platform = ? AND forge_user_id = ?", "github", ghUser.ID).Get(&mapping); err == nil {
				if timeutil.TimeStampNow()-mapping.LastSyncedUnix < 24*3600 {
					log.Trace("ForgeBridge: Sync for GitHub User %s skipped (already synced within 24h)", githubUsername)
					return nil
				}
			}
		}

		// Update Gitea User if fields are empty
		giteaUser, err := user_model.GetUserByID(ctx, task.RepoOwnerID)
		if err != nil {
			if user_model.IsErrUserNotExist(err) {
				return nil // user got deleted?
			}
			return err
		}

		var colsToUpdate []string
		if giteaUser.FullName == "" && ghUser.Name != "" {
			giteaUser.FullName = ghUser.Name
			colsToUpdate = append(colsToUpdate, "full_name")
		}
		if giteaUser.Description == "" && ghUser.Bio != "" {
			giteaUser.Description = ghUser.Bio
			colsToUpdate = append(colsToUpdate, "description")
		}
		if giteaUser.Location == "" && ghUser.Location != "" {
			giteaUser.Location = ghUser.Location
			colsToUpdate = append(colsToUpdate, "location")
		}
		if giteaUser.Website == "" && ghUser.Blog != "" {
			giteaUser.Website = ghUser.Blog
			colsToUpdate = append(colsToUpdate, "website")
		}
		// TODO: handle avatar logic if current is default

		if len(colsToUpdate) > 0 {
			if err := user_model.UpdateUserCols(ctx, giteaUser, colsToUpdate...); err != nil {
				return fmt.Errorf("failed to update user %d: %v", giteaUser.ID, err)
			}
		}

		// Update or Insert Mapping
		mapping := &forgebridge.ForgeUserMapping{
			Platform:       "github",
			GiteaUserID:    giteaUser.ID,
			ForgeUserID:    ghUser.ID,
			ForgeLogin:     ghUser.Login,
			LastSyncedUnix: timeutil.TimeStampNow(),
		}
		
		if exists {
			_, err = db.GetEngine(ctx).Where("platform = ? AND forge_user_id = ?", "github", ghUser.ID).
				Cols("gitea_user_id", "forge_login", "last_synced_unix").
				Update(mapping)
		} else {
			_, err = db.GetEngine(ctx).Insert(mapping)
		}

		return err
	})
}

// fetchAndMapGitHubUser fetches data for a specific mapping task triggered by UC8 (Mapping)
func fetchAndMapGitHubUser(ctx context.Context, task *MappingTask) error {
	// Provide delay for general mapping items since we have lots of them
	// Start with trying to get a token
	tokenStr, tokenSource, err := GetTokenForUser(ctx, task.RepoOwnerID, "github")
	if err != nil && err != ErrNoTokensAvailable {
		log.Warn("ForgeBridge: GetTokenForUser error mapping %s: %v", task.OriginalAuthor, err)
	}
	log.Trace("ForgeBridge: Using token from %s for mapping github author %s", tokenSource, task.OriginalAuthor)

	reqTarget := fmt.Sprintf("https://api.github.com/users/%s", task.OriginalAuthor)
	req, err := http.NewRequestWithContext(ctx, "GET", reqTarget, nil)
	if err != nil {
		return err
	}
	if tokenStr != "" {
		req.Header.Set("Authorization", "Bearer "+tokenStr)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse Rate Limit & Delay
	rateLimit := ParseRateLimitHeaders(resp.Header)
	delay := CalculateAdaptiveDelay(rateLimit, 2*time.Second) // 2 sec base delay for background mappings
	if delay > 0 {
		log.Trace("ForgeBridge: Rate limit aware delay: %v", delay)
		time.Sleep(delay)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("github api returned status %d for user %s", resp.StatusCode, task.OriginalAuthor)
	}

	var ghUser GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return err
	}

	// Insert into Mapping Table (Phase 8 completion)
	// We don't link to Gitea user right now because Gitea user creation is another workflow, 
	// but we fetch the true GithubUserID which allows deduplication.
	return forgebridge.InsertGithubUserMapping(ctx, ghUser.ID, 0, ghUser.Login)
}

// === END CUSTOM: forge-bridge ===
