// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"

	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/modules/log"
)

// ProcessUserMapping Drains the SyncTask queue to fetch GitHub user ID -> Gitea User ID connections
// This runs sequentially as part of the ForgeBridge Sync Engine
func ProcessUserMapping(ctx context.Context, task *SyncTask) error {
	repo, err := repo_model.GetRepositoryByID(ctx, task.RepoID)
	if err != nil {
		log.Error("ForgeBridge ProcessUserMapping: failed to get repo %d: %v", task.RepoID, err)
		return err
	}

	log.Info("ForgeBridge ProcessUserMapping: Beginning deduplication for %s", repo.FullName())

	// TODO: Look up external ID of users associated with this repository
	// (commit authors, reviewers, assignees) by leveraging the Gitea Migration data
	// If a matching GitHub User ID is found, map it in models.GithubUserMapping

	return nil
}

// === END CUSTOM: forge-bridge ===
