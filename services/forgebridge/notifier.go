// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"

	repo_model "code.gitea.io/gitea/models/repo"
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/log"
	notify_service "code.gitea.io/gitea/services/notify"
)

// Notifier implements notify.Notifier to capture repository migration events
type Notifier struct {
	notify_service.NullNotifier
}

var _ notify_service.Notifier = &Notifier{}

// NewNotifier creates a new Notifier for Forge Bridge
func NewNotifier() *Notifier {
	return &Notifier{}
}

// MigrateRepository captures the event when a repository is successfully migrated
func (n *Notifier) MigrateRepository(ctx context.Context, doer, u *user_model.User, repo *repo_model.Repository) {
	if repo.OriginalURL == "" {
		return
	}

	log.Trace("ForgeBridge Notifier: captured MigrateRepository event for RepoID %d (Owner: %s)", repo.ID, repo.OwnerName)

	// Post-migration hooks for ForgeBridge User Mapping and deduplication (Phase 8 logic)
	// Triggers a background job to sync missing user info from GitHub to Gitea mapping table.
	log.Info("ForgeBridge MigrationInterceptor: Background task triggered for mapping logic on repo %v", repo.ID)

	task := &SyncTask{
		RepoID:      repo.ID,
		RepoOwnerID: repo.OwnerID,
		OriginalURL: repo.OriginalURL,
	}

	PushSyncTask(task)
}

// === END CUSTOM: forge-bridge ===
