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


func GithubCallback(c *gin.Context) {
	code := c.Query("code")
	echoed := c.Query("state")
	envelope := "oauth_state"

	_,err := service.VerifyState(echoed,envelope,"github")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	



	c.JSON(http.StatusOK, gin.H{"message": "GitHub callback received", "code": code})

}

func GoogleCallback(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Google callback received"})
}