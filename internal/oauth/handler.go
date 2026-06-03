package oauth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"

	accountsvc "deploy-hub/internal/accounts/service"
	"deploy-hub/internal/oauth/service"
)

func setStateCookie(c *gin.Context, envelope string) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := os.Getenv("OAUTH_STATE_COOKIE_SECURE") == "true"
	c.SetCookie("oauth_state", envelope, service.StateTTLSeconds(), "/", "", secure, true)
}

func clearStateCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := os.Getenv("OAUTH_STATE_COOKIE_SECURE") == "true"
	c.SetCookie("oauth_state", "", -1, "/", "", secure, true)
}

func spaRedirect(c *gin.Context, query map[string]string) {
	values := url.Values{}
	for key, value := range query {
		values.Set(key, value)
	}
	c.Redirect(http.StatusFound, os.Getenv("SPA_AUTH_COMPLETE_URL")+"#"+values.Encode())
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

func handleCallback(c *gin.Context, provider string, exchangeAndResolve func(ctx context.Context, code string) (ResolutionResult, error)) {
	code := c.Query("code")
	echoed := c.Query("state")
	envelope, _ := c.Cookie("oauth_state")

	if _, err := service.VerifyState(echoed, envelope, provider); err != nil {
		clearStateCookie(c)
		spaRedirect(c, map[string]string{"error": "oauth_state_invalid", "message": err.Error()})
		return
	}

	if code == "" {
		clearStateCookie(c)
		spaRedirect(c, map[string]string{"error": "oauth_no_code", "message": "authorization code missing"})
		return
	}

	result, err := exchangeAndResolve(c.Request.Context(), code)
	if err != nil {
		clearStateCookie(c)
		if errors.Is(err, ErrNoVerifiedEmail) {
			spaRedirect(c, map[string]string{"error": "oauth_error", "message": err.Error()})
		} else {
			spaRedirect(c, map[string]string{"error": "oauth_provider_error", "message": err.Error()})
		}
		return
	}

	pair, err := accountsvc.IssueJWTPair(result.User)
	if err != nil {
		clearStateCookie(c)
		spaRedirect(c, map[string]string{"error": "oauth_error", "message": "failed to issue tokens"})
		return
	}

	accountsvc.SetRefreshCookie(c.Writer, pair.Refresh)
	clearStateCookie(c)
	spaRedirect(c, map[string]string{"access": pair.Access, "intent": "login"})
}

func GitHubStart(c *gin.Context) { startOAuth(c, "github", service.BuildGitHubAuthorizeURL) }

func GoogleStart(c *gin.Context) { startOAuth(c, "google", service.BuildGoogleAuthorizeURL) }

func GithubCallback(c *gin.Context) {
	handleCallback(c, "github", func(ctx context.Context, code string) (ResolutionResult, error) {
		tokens, err := service.ExchangeGitHubCode(ctx, code)
		if err != nil {
			return ResolutionResult{}, err
		}
		identity, err := service.FetchGitHubIdentity(ctx, tokens.AccessToken)
		if err != nil {
			return ResolutionResult{}, err
		}
		return ResolveGitHub(*identity, *tokens)
	})
}

func GoogleCallback(c *gin.Context) {
	handleCallback(c, "google", func(ctx context.Context, code string) (ResolutionResult, error) {
		tokens, err := service.ExchangeGoogleCode(ctx, code)
		if err != nil {
			return ResolutionResult{}, err
		}
		identity, err := service.VerifyGoogleIDToken(ctx, tokens.IDToken)
		if err != nil {
			return ResolutionResult{}, err
		}
		return ResolveGoogle(*identity, *tokens)
	})
}
