package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const githubAPIBase = "https://api.github.com"

var githubAPIClient = &http.Client{Timeout: 15 * time.Second}

type githubRepo struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	FullName      string  `json:"full_name"`
	Private       bool    `json:"private"`
	DefaultBranch string  `json:"default_branch"`
	Description   string  `json:"description"`
	HTMLURL       string  `json:"html_url"`
	CreatedAt     *string `json:"created_at"`
	PushedAt      *string `json:"pushed_at"`
}

// fetchUserRepos pages through GET /user/repos until a short page is returned
// or maxPages is hit (whichever comes first).
func fetchUserRepos(ctx context.Context, accessToken string) ([]githubRepo, error) {
	const perPage = 100
	const maxPages = 10
	var all []githubRepo

	for page := 1; page <= maxPages; page++ {
		params := url.Values{
			"per_page":    {strconv.Itoa(perPage)},
			"page":        {strconv.Itoa(page)},
			"sort":        {"pushed"},
			"direction":   {"desc"},
			"affiliation": {"owner,collaborator,organization_member"},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBase+"/user/repos?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		req.Header.Set("User-Agent", "deploy-hub-backend")

		resp, err := githubAPIClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("github api %d: %s", resp.StatusCode, body)
		}

		var pageRepos []githubRepo
		if err := json.Unmarshal(body, &pageRepos); err != nil {
			return nil, err
		}

		all = append(all, pageRepos...)
		if len(pageRepos) < perPage {
			break
		}
	}
	return all, nil
}
