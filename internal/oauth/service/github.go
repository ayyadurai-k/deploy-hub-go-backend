package service

import (
	"net/url"
	"os"
	"strings"
)

const (
	GITHUB_AUTHORIZE_URL = "https://github.com/login/oauth/authorize"
	GITHUB_TOKEN_URL     = "https://github.com/login/oauth/access_token"
)

var (
	GITHUB_OAUTH_CLIENT_ID     = os.Getenv("GITHUB_OAUTH_CLIENT_ID")
	GITHUB_OAUTH_CLIENT_SECRET = os.Getenv("GITHUB_OAUTH_CLIENT_SECRET")
	GITHUB_OAUTH_REDIRECT_URI  = os.Getenv("GITHUB_OAUTH_REDIRECT_URI")
	GITHUB_SCOPES              = []string{"read:user", "user:email", "repo"}
)

type GitHubTokens struct {
	AccessToken string
}

type GitHubIdentity struct {
	UserID    int64
	Login     string
	Name      string
	Email     string
	AvatarURL string
}

func BuildGitHubAuthorizeURL(state string) string {
	params := url.Values{
		"client_id":    []string{GITHUB_OAUTH_CLIENT_ID},
		"redirect_uri": []string{GITHUB_OAUTH_REDIRECT_URI},
		"scope":        []string{strings.Join(GITHUB_SCOPES, " ")},
		"state":        []string{state},
		"allow_signup": []string{"true"},
	}

	return GITHUB_AUTHORIZE_URL + "?" + params.Encode()
}
