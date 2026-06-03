package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var githubAPIClient = &http.Client{Timeout: 10 * time.Second}

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

func githubOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URI"),
		Scopes:       []string{"read:user", "user:email", "repo"},
		Endpoint:     github.Endpoint,
	}
}

func BuildGitHubAuthorizeURL(state string) string {
	return githubOAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("allow_signup", "true"))
}

func ExchangeGitHubCode(ctx context.Context, code string) (*GitHubTokens, error) {
	token, err := githubOAuthConfig().Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return &GitHubTokens{AccessToken: token.AccessToken}, nil
}

func FetchGitHubIdentity(ctx context.Context, accessToken string) (*GitHubIdentity, error) {
	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := githubGetJSON(ctx, "https://api.github.com/user", accessToken, &user); err != nil {
		return nil, fmt.Errorf("/user fetch failed: %w", err)
	}

	email := user.Email
	if email == "" {
		email = githubPrimaryEmail(ctx, accessToken)
	}

	return &GitHubIdentity{
		UserID:    user.ID,
		Login:     user.Login,
		Name:      user.Name,
		Email:     email,
		AvatarURL: user.AvatarURL,
	}, nil
}

func githubPrimaryEmail(ctx context.Context, accessToken string) string {
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := githubGetJSON(ctx, "https://api.github.com/user/emails", accessToken, &emails); err != nil {
		return ""
	}
	for _, entry := range emails {
		if entry.Primary && entry.Verified {
			return entry.Email
		}
	}
	return ""
}

func githubGetJSON(ctx context.Context, url, accessToken string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "repo-manage-backend")

	resp, err := githubAPIClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%d %s", resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
