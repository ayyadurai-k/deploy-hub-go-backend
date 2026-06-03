package oauth

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"deploy-hub/internal/oauth/service"
)

func setStateCookie(c *gin.Context, envelope string) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := os.Getenv("OAUTH_STATE_COOKIE_SECURE") == "true"
	c.SetCookie("oauth_state", envelope, service.StateTTLSeconds(), "/", "", secure, true)
}

func startOAuth(c *gin.Context, provider string, buildAuthorizeURL func(string) string) {
	nonce, envelope, err := service.IssueState(provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue state"})
		return
	}

	setStateCookie(c, envelope)
	c.Redirect(http.StatusFound, buildAuthorizeURL(nonce))
}

func GitHubStart(c *gin.Context) { startOAuth(c, "github", service.BuildGitHubAuthorizeURL) }

func GoogleStart(c *gin.Context) { startOAuth(c, "google", service.BuildGoogleAuthorizeURL) }
