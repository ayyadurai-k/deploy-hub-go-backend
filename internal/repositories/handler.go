package repositories

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"deploy-hub/config"
	"deploy-hub/internal/accounts"
	"deploy-hub/internal/httpx"
	"deploy-hub/internal/oauth"
)

func currentUser(c *gin.Context) accounts.User {
	return c.MustGet(accounts.CtxUserKey).(accounts.User)
}

// requireGitHubProfile loads the GitHub profile for a user, or reports that
// the user hasn't connected GitHub yet.
func requireGitHubProfile(userID uint) (*oauth.GitHubProfile, bool) {
	var profile oauth.GitHubProfile
	if err := config.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, false
	}
	return &profile, true
}

func List(c *gin.Context) {
	profile, ok := requireGitHubProfile(currentUser(c).ID)
	if !ok {
		httpx.Error(c, http.StatusConflict, "github_not_connected", "GitHub account not connected")
		return
	}

	var repos []Repository
	config.DB.Where("github_profile_id = ?", profile.ID).
		Order("github_pushed_at DESC NULLS LAST, id").
		Find(&repos)

	results := make([]RepositoryResponse, 0, len(repos))
	for _, repo := range repos {
		results = append(results, ToRepositoryResponse(repo))
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

func Sync(c *gin.Context) {
	profile, ok := requireGitHubProfile(currentUser(c).ID)
	if !ok {
		httpx.Error(c, http.StatusConflict, "github_not_connected", "GitHub account not connected")
		return
	}

	count, err := SyncRepositories(c.Request.Context(), profile)
	if err != nil {
		httpx.Error(c, http.StatusBadGateway, "github_api_error", "failed to fetch repositories from GitHub")
		return
	}
	c.JSON(http.StatusOK, gin.H{"synced": count, "status": profile.LastSyncStatus})
}
