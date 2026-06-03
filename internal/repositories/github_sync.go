package repositories

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"deploy-hub/config"
	"deploy-hub/internal/oauth"
)

// SyncRepositories pulls all repos for the profile and upserts them in one
// statement (INSERT ... ON CONFLICT DO UPDATE), then records the sync status
// on the profile. Returns the number of repos seen.
func SyncRepositories(ctx context.Context, profile *oauth.GitHubProfile) (int, error) {
	config.DB.Model(profile).Updates(map[string]any{
		"last_sync_status": oauth.SyncInProgress,
		"last_sync_error":  "",
	})

	accessToken, err := profile.GetAccessToken()
	if err != nil {
		markSyncFailure(profile, err)
		return 0, err
	}

	rawRepos, err := fetchUserRepos(ctx, accessToken)
	if err != nil {
		markSyncFailure(profile, err)
		return 0, err
	}

	rows := make([]Repository, 0, len(rawRepos))
	for _, raw := range rawRepos {
		rows = append(rows, Repository{
			GitHubProfileID: profile.ID,
			GitHubRepoID:    raw.ID,
			Name:            raw.Name,
			FullName:        raw.FullName,
			Private:         raw.Private,
			DefaultBranch:   raw.DefaultBranch,
			Description:     raw.Description,
			HTMLURL:         raw.HTMLURL,
			GitHubCreatedAt: parseGitHubTime(raw.CreatedAt),
			GitHubPushedAt:  parseGitHubTime(raw.PushedAt),
		})
	}

	if len(rows) > 0 {
		err = config.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "github_profile_id"}, {Name: "github_repo_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "full_name", "private", "default_branch",
				"description", "html_url", "github_created_at", "github_pushed_at", "updated_at",
			}),
		}).Create(&rows).Error
		if err != nil {
			markSyncFailure(profile, err)
			return 0, err
		}
	}

	now := time.Now()
	config.DB.Model(profile).Updates(map[string]any{
		"last_sync_status": oauth.SyncSuccess,
		"last_sync_error":  "",
		"last_synced_at":   &now,
	})
	return len(rows), nil
}

func markSyncFailure(profile *oauth.GitHubProfile, cause error) {
	message := cause.Error()
	if len(message) > 5000 {
		message = message[:5000]
	}
	now := time.Now()
	config.DB.Model(profile).Updates(map[string]any{
		"last_sync_status": oauth.SyncFailure,
		"last_sync_error":  message,
		"last_synced_at":   &now,
	})
}

func parseGitHubTime(value *string) *time.Time {
	if value == nil || *value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil
	}
	return &parsed
}
