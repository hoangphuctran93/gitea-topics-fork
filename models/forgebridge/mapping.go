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
// === END CUSTOM: forge-bridge ===
