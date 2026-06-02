package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubUserURL      = "https://api.github.com/user"
	githubEmailsURL    = "https://api.github.com/user/emails"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

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
		"client_id":    {os.Getenv("GITHUB_CLIENT_ID")},
		"redirect_uri": {os.Getenv("GITHUB_REDIRECT_URI")},
		"scope":        {os.Getenv("GITHUB_SCOPES")},
		"state":        {state},
		"allow_signup": {"true"},
	}
	return githubAuthorizeURL + "?" + params.Encode()
}

func ExchangeGitHubCode(code string) (GitHubTokens, error) {
	form := url.Values{
		"client_id":     {os.Getenv("GITHUB_CLIENT_ID")},
		"client_secret": {os.Getenv("GITHUB_CLIENT_SECRET")},
		"code":          {code},
		"redirect_uri":  {os.Getenv("GITHUB_REDIRECT_URI")},
	}

	req, err := http.NewRequest(http.MethodPost, githubTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return GitHubTokens{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return GitHubTokens{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return GitHubTokens{}, fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, body)
	}

	var body struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return GitHubTokens{}, err
	}
	if body.Error != "" {
		message := body.ErrorDescription
		if message == "" {
			message = body.Error
		}
		return GitHubTokens{}, fmt.Errorf("token exchange error: %s", message)
	}
	return GitHubTokens{AccessToken: body.AccessToken}, nil
}

func FetchGitHubIdentity(accessToken string) (GitHubIdentity, error) {
	resp, err := githubGet(githubUserURL, accessToken)
	if err != nil {
		return GitHubIdentity{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return GitHubIdentity{}, fmt.Errorf("/user fetch failed: %d %s", resp.StatusCode, body)
	}

	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return GitHubIdentity{}, err
	}

	email := user.Email
	if email == "" {
		email = githubPrimaryEmail(accessToken)
	}

	return GitHubIdentity{
		UserID:    user.ID,
		Login:     user.Login,
		Name:      user.Name,
		Email:     email,
		AvatarURL: user.AvatarURL,
	}, nil
}

func githubPrimaryEmail(accessToken string) string {
	resp, err := githubGet(githubEmailsURL, accessToken)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return ""
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}
	for _, entry := range emails {
		if entry.Primary && entry.Verified {
			return entry.Email
		}
	}
	return ""
}

func githubGet(targetURL, accessToken string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "repo-manage-backend")
	return httpClient.Do(req)
}
