package repositories

import (
	"time"

	"deploy-hub/internal/oauth"
)

type Repository struct {
	ID              uint                `json:"id" gorm:"primarykey"`
	GitHubProfileID uint                `json:"github_profile_id" gorm:"column:github_profile_id;uniqueIndex:unique_repo_per_github_profile;not null"`
	GitHubProfile   oauth.GitHubProfile `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	GitHubRepoID    int64               `json:"github_repo_id" gorm:"column:github_repo_id;uniqueIndex:unique_repo_per_github_profile;not null"`
	Name            string              `json:"name" gorm:"column:name;not null"`
	FullName        string              `json:"full_name" gorm:"column:full_name;not null"`
	Private         bool                `json:"private" gorm:"column:private;default:false"`
	DefaultBranch   string              `json:"default_branch" gorm:"column:default_branch"`
	Description     string              `json:"description" gorm:"column:description"`
	HTMLURL         string              `json:"html_url" gorm:"column:html_url;not null"`
	GitHubCreatedAt *time.Time          `json:"github_created_at" gorm:"column:github_created_at"`
	GitHubPushedAt  *time.Time          `json:"github_pushed_at" gorm:"column:github_pushed_at"`
	CreatedAt       time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
}