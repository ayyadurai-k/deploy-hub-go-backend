package repositories

import (
	"net/http"
	"net/url"
	"strconv"

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

	limit, offset := paginationParams(c)

	var count int64
	config.DB.Model(&Repository{}).Where("github_profile_id = ?", profile.ID).Count(&count)

	var repos []Repository
	config.DB.Where("github_profile_id = ?", profile.ID).
		Order("github_pushed_at DESC NULLS LAST, id").
		Limit(limit).Offset(offset).
		Find(&repos)

	results := make([]RepositoryResponse, 0, len(repos))
	for _, repo := range repos {
		results = append(results, ToRepositoryResponse(repo))
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    count,
		"next":     pageURL(c, offset+limit, limit, count),
		"previous": pageURL(c, offset-limit, limit, count),
		"results":  results,
	})
}

// paginationParams mirrors DRF LimitOffsetPagination: default limit 10, max 100.
func paginationParams(c *gin.Context) (limit, offset int) {
	limit = 10
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if limit > 100 {
		limit = 100
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v > 0 {
		offset = v
	}
	return limit, offset
}

// pageURL builds the next/previous link, or nil when that page is out of range.
func pageURL(c *gin.Context, offset, limit int, count int64) *string {
	if offset < 0 || int64(offset) >= count {
		return nil
	}
	values := url.Values{"limit": {strconv.Itoa(limit)}}
	if offset > 0 {
		values.Set("offset", strconv.Itoa(offset))
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	full := scheme + "://" + c.Request.Host + c.Request.URL.Path + "?" + values.Encode()
	return &full
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
