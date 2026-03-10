// === BEGIN CUSTOM: forge-bridge ===
// Upstream-safe: false | Author: hoangphuctran93

package forgebridge

import (
	"context"

	forgebridge_model "code.gitea.io/gitea/models/forgebridge"
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

	log.Info("ForgeBridge ProcessUserMapping: Beginning deduplication metadata extraction for %s", repo.FullName())

	// 1. Fetch distinct OriginalAuthors from Local DB (Issues/Comments/Reviews)
	mappings, err := forgebridge_model.GetOriginalAuthorsFromRepository(ctx, task.RepoID)
	if err != nil {
		log.Error("ForgeBridge ProcessUserMapping: failed to extract local metadata for repo %d: %v", task.RepoID, err)
		return err
	}

	if len(mappings) == 0 {
		log.Info("ForgeBridge ProcessUserMapping: No external metadata found for %s", repo.FullName())
		return nil
	}

	log.Info("ForgeBridge ProcessUserMapping: Found %d unique external authors in %s. Pushing to rate-limited queue...", len(mappings), repo.FullName())

	// 2. Push to delayed queue instead of batch insert local
	for _, mapping := range mappings {
		PushMappingTask(&MappingTask{
			OriginalAuthor:   mapping.GithubUsername,
			OriginalAuthorID: mapping.GithubUserID,
			RepoOwnerID:      task.RepoOwnerID,
		})
	}

	log.Info("ForgeBridge ProcessUserMapping: Successfully queued %d users for %s", len(mappings), repo.FullName())

	return nil
}

// === END CUSTOM: forge-bridge ===
