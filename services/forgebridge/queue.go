// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"
	"time"

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

// MappingTask represents a user mapping fetch task protected by rate limits
type MappingTask struct {
	OriginalAuthor   string
	OriginalAuthorID int64
	RepoOwnerID      int64
	RetryCount       int
}

var (
	syncQueue    *queue.WorkerPoolQueue[*SyncTask]
	mappingQueue *queue.WorkerPoolQueue[*MappingTask]
)

// InitQueue initializes the background queue for Forge Bridge tasks
func InitQueue(ctx context.Context) {
	if syncQueue == nil {
		syncQueue = queue.CreateSimpleQueue(graceful.GetManager().ShutdownContext(), "forge_bridge_sync", handler)
		if syncQueue == nil {
			log.Fatal("Unable to create forge_bridge_sync queue")
		}
		go graceful.GetManager().RunWithCancel(syncQueue)
	}

	if mappingQueue == nil {
		mappingQueue = queue.CreateSimpleQueue(graceful.GetManager().ShutdownContext(), "forge_bridge_mapping", mappingHandler)
		if mappingQueue == nil {
			log.Fatal("Unable to create forge_bridge_mapping queue")
		}
		go graceful.GetManager().RunWithCancel(mappingQueue)
	}
}

func handler(items ...*SyncTask) []*SyncTask {
	for _, task := range items {
		log.Trace("Processing ForgeBridge SyncTask for RepoID: %d", task.RepoID)
		if err := syncGitHubProfile(graceful.GetManager().ShutdownContext(), task); err != nil {
			log.Error("Failed to process ForgeBridge Profile Sync for RepoID %d: %v", task.RepoID, err)
		} else {
			log.Trace("Successfully processed ForgeBridge Profile Sync for RepoID %d", task.RepoID)
		}

		// Phase 8 & 11: Trigger User Mapping & pushing into delayed queue
		if err := ProcessUserMapping(graceful.GetManager().ShutdownContext(), task); err != nil {
			log.Error("Failed to process ForgeBridge User Mapping for RepoID %d: %v", task.RepoID, err)
		} else {
			log.Trace("Successfully initiated ForgeBridge User Mapping for RepoID %d", task.RepoID)
		}
	}
	return nil
}

func mappingHandler(items ...*MappingTask) []*MappingTask {
	for _, task := range items {
		log.Trace("Processing ForgeBridge MappingTask for user: %s", task.OriginalAuthor)
		if err := fetchAndMapGitHubUser(graceful.GetManager().ShutdownContext(), task); err != nil {
			log.Error("Failed to process ForgeBridge Mapping for user %s: %v", task.OriginalAuthor, err)
			if task.RetryCount < 3 {
				task.RetryCount++
				log.Warn("Retrying ForgeBridge Mapping for %s (Attempt %d/3)", task.OriginalAuthor, task.RetryCount)
				// Small penalty before re-queueing to avoid thrashing.
				// Respects shutdown context so the goroutine exits cleanly on server stop.
				go func(t *MappingTask) {
					shutdownCtx := graceful.GetManager().ShutdownContext()
					select {
					case <-shutdownCtx.Done():
						log.Warn("ForgeBridge: Retry for %s cancelled due to shutdown", t.OriginalAuthor)
						return
					case <-time.After(5 * time.Second):
						PushMappingTask(t)
					}
				}(task)
			} else {
				log.Error("Max retries exceeded for ForgeBridge Mapping %s. Dropping.", task.OriginalAuthor)
			}
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
		if err := syncQueue.Push(task); err != nil {
			log.Error("ForgeBridge: failed to push SyncTask for RepoID %d: %v", task.RepoID, err)
		}
	}()
}

// PushMappingTask asynchronously pushes a user mapping task to the delayed queue
func PushMappingTask(task *MappingTask) {
	if mappingQueue == nil {
		log.Error("ForgeBridge: PushMappingTask invoked but mappingQueue is not initialized")
		return
	}

	go func() {
		if err := mappingQueue.Push(task); err != nil {
			log.Error("ForgeBridge: failed to push MappingTask for user %s: %v", task.OriginalAuthor, err)
		}
	}()
}

// === END CUSTOM: forge-bridge ===
