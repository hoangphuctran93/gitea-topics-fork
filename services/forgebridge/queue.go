// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"

	"code.gitea.io/gitea/modules/graceful"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/queue"
)

// SyncTask represents a background task for Forge Bridge (UC2/UC3)
type SyncTask struct {
	RepoID      int64
	RepoOwnerID int64
	OriginalURL string
}

var syncQueue *queue.WorkerPoolQueue[*SyncTask]

// InitQueue initializes the background queue for Forge Bridge tasks
func InitQueue(ctx context.Context) {
	if syncQueue != nil {
		return
	}

	syncQueue = queue.CreateSimpleQueue(graceful.GetManager().ShutdownContext(), "forge_bridge_sync", handler)
	if syncQueue == nil {
		log.Fatal("Unable to create forge_bridge_sync queue")
	}
	go graceful.GetManager().RunWithCancel(syncQueue)
}

func handler(items ...*SyncTask) []*SyncTask {
	for _, task := range items {
		log.Trace("Processing ForgeBridge SyncTask for RepoID: %d", task.RepoID)
		if err := syncGitHubProfile(graceful.GetManager().ShutdownContext(), task); err != nil {
			log.Error("Failed to process ForgeBridge Profile Sync for RepoID %d: %v", task.RepoID, err)
		} else {
			log.Trace("Successfully processed ForgeBridge Profile Sync for RepoID %d", task.RepoID)
		}

		// Phase 8: Trigger User Mapping & Deduplication logic
		if err := ProcessUserMapping(graceful.GetManager().ShutdownContext(), task); err != nil {
			log.Error("Failed to process ForgeBridge User Mapping for RepoID %d: %v", task.RepoID, err)
		} else {
			log.Trace("Successfully processed ForgeBridge User Mapping for RepoID %d", task.RepoID)
		}
	}
	return nil
}

// PushSyncTask asynchronously pushes a task to the queue
func PushSyncTask(task *SyncTask) {
	if syncQueue == nil {
		log.Error("ForgeBridge: PushSyncTask invoked but syncQueue is not initialized")
		return
	}

	go func() {
		_ = syncQueue.Push(task)
	}()
}

// === END CUSTOM: forge-bridge ===
