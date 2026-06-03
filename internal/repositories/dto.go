package repositories

import "time"

type RepositoryResponse struct {
	ID             uint       `json:"id"`
	GitHubRepoID   int64      `json:"github_repo_id"`
	Name           string     `json:"name"`
	FullName       string     `json:"full_name"`
	Private        bool       `json:"private"`
	DefaultBranch  string     `json:"default_branch"`
	Description    string     `json:"description"`
	HTMLURL        string     `json:"html_url"`
	GitHubPushedAt *time.Time `json:"github_pushed_at"`
}

func ToRepositoryResponse(r Repository) RepositoryResponse {
	return RepositoryResponse{
		ID:             r.ID,
		GitHubRepoID:   r.GitHubRepoID,
		Name:           r.Name,
		FullName:       r.FullName,
		Private:        r.Private,
		DefaultBranch:  r.DefaultBranch,
		Description:    r.Description,
		HTMLURL:        r.HTMLURL,
		GitHubPushedAt: r.GitHubPushedAt,
	}
}
