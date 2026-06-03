package oauth

import (
	"deploy-hub/internal/oauth/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// utils

func setCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", true, true)
}


func GitHubStart(c *gin.Context){
	provider := "github"
	nonce, envelope, err := service.IssueState(provider)

	if err != nil{
		c.JSON(500,gin.H{
			"error":"failed to issue state",
		})
		return
	}

	authorizeURL := service.BuildGitHubAuthorizeURL(nonce)
	setCookie(c, "oauth_state", envelope, 600)
	c.Redirect(http.StatusFound,authorizeURL)
}

func GoogleStart(c *gin.Context){
	provider := "google"
	nonce, envelope, err := service.IssueState(provider)

	if err != nil{
		c.JSON(500,gin.H{
			"error":"failed to issue state",
		})
		return
	}

	authorizeURL := service.BuildGoogleAuthorizeURL(nonce)
	setCookie(c, "oauth_state", envelope, 600)
	c.Redirect(http.StatusFound,authorizeURL)
}