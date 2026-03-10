// === BEGIN CUSTOM: forge-bridge ===
package forgebridge

import (
	"context"

	"code.gitea.io/gitea/models/db"
	user_model "code.gitea.io/gitea/models/user"
)

// GithubUserMapping maps a remote GitHub user ID to a local Gitea user ID.
// This ensures that when migrating multiple repositories involving the same GitHub user,
// Gitea can reuse the existing local user/organization rather than creating duplicates.
// Dữ liệu gốc trên Forge (GitHub) được tôn trọng, và liên kết được duy trì qua bảng này.
type GithubUserMapping struct {
	ID             int64 `xorm:"pk autoincr"`
	GithubUserID   int64 `xorm:"UNIQUE INDEX NOT NULL"` // ID thực trên GitHub
	GiteaUserID    int64 `xorm:"INDEX NOT NULL"`        // ID user nội bộ trên Gitea
	GithubUsername string
	CreatedUnix    int64 `xorm:"created"`
	UpdatedUnix    int64 `xorm:"updated"`
}

func init() {
	db.RegisterModel(new(GithubUserMapping))
}

// GetGiteaUserByGithubID retrieves the local Gitea User mapped to the given GitHub User ID.
func GetGiteaUserByGithubID(ctx context.Context, githubID int64) (*user_model.User, error) {
	mapping := new(GithubUserMapping)
	has, err := db.GetEngine(ctx).Where("github_user_id = ?", githubID).Get(mapping)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, user_model.ErrUserNotExist{
			UID:   0,
			Name:  "",
		}
	}

	return user_model.GetUserByID(ctx, mapping.GiteaUserID)
}

// InsertGithubUserMapping creates a new mapping between a GitHub User ID and a Gitea User ID.
func InsertGithubUserMapping(ctx context.Context, githubID int64, giteaID int64, githubUsername string) error {
	mapping := &GithubUserMapping{
		GithubUserID:   githubID,
		GiteaUserID:    giteaID,
		GithubUsername: githubUsername,
	}
	return db.Insert(ctx, mapping)
}

// BatchInsertUserMappings efficiently inserts multiple GithubUserMapping records from local metadata, ignoring duplicates.
// This is used during the Background Metadata Mapper phase to populate the table quickly from local Issues/Comments.
func BatchInsertUserMappings(ctx context.Context, mappings []*GithubUserMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	return db.WithTx(ctx, func(ctx context.Context) error {
		for _, m := range mappings {
			if m.GithubUserID <= 0 {
				continue // Skip invalid IDs
			}
			has, err := db.GetEngine(ctx).Where("github_user_id = ?", m.GithubUserID).Get(new(GithubUserMapping))
			if err != nil {
				return err
			}
			if !has {
				if err := db.Insert(ctx, m); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// GetOriginalAuthorsFromRepository extracts distinct OriginalAuthorID and OriginalAuthor
// stored in the local database (Issues and Comments) after a repository migration.
func GetOriginalAuthorsFromRepository(ctx context.Context, repoID int64) ([]*GithubUserMapping, error) {
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
		// Ignore if table doesn't exist or other minor issues, but usually it should be there.
		return nil, err
	}

	// Deduplicate in memory
	seen := make(map[int64]bool)
	var mappings []*GithubUserMapping

	allAuthors := append(issueAuthors, commentAuthors...)
	allAuthors = append(allAuthors, reviewAuthors...)

	for _, a := range allAuthors {
		if !seen[a.OriginalAuthorID] {
			seen[a.OriginalAuthorID] = true
			mappings = append(mappings, &GithubUserMapping{
				GithubUserID:   a.OriginalAuthorID,
				GithubUsername: a.OriginalAuthor,
				GiteaUserID:    0, // Will be mapped later in Phase 10
			})
		}
	}

	return mappings, nil
}
// === END CUSTOM: forge-bridge ===
