package oauth

import (
	"time"

	"deploy-hub/internal/accounts"
	"deploy-hub/internal/oauth/service"
)

type SyncStatus string

const (
	SyncPending    SyncStatus = "pending"
	SyncInProgress SyncStatus = "in_progress"
	SyncSuccess    SyncStatus = "success"
	SyncFailure    SyncStatus = "failure"
)

type OAuthProfile struct {
	ID                   uint      `json:"id" gorm:"primarykey"`
	AccessTokenEncrypted string    `json:"-" gorm:"column:access_token_encrypted;not null"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (p *OAuthProfile) SetAccessToken(plaintext string) error {
	ciphertext, err := service.Encrypt(plaintext)
	if err != nil {
		return err
	}
	p.AccessTokenEncrypted = ciphertext
	return nil
}

func (p *OAuthProfile) GetAccessToken() (string, error) {
	return service.Decrypt(p.AccessTokenEncrypted)
}

type GoogleProfile struct {
	OAuthProfile
	UserID     uint          `json:"user_id" gorm:"uniqueIndex;not null"`
	User       accounts.User `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	GoogleSub  string        `json:"google_sub" gorm:"column:google_sub;uniqueIndex;not null"`
	Email      string        `json:"email" gorm:"column:email;not null"`
	PictureURL string        `json:"picture_url" gorm:"column:picture_url"`
}


type GitHubProfile struct {
	OAuthProfile
	UserID         uint          `json:"user_id" gorm:"uniqueIndex;not null"`
	User           accounts.User `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	GitHubUserID   int64         `json:"github_user_id" gorm:"column:github_user_id;uniqueIndex;not null"`
	GitHubLogin    string        `json:"github_login" gorm:"column:github_login;not null"`
	AvatarURL      string        `json:"avatar_url" gorm:"column:avatar_url"`
	LastSyncedAt   *time.Time    `json:"last_synced_at" gorm:"column:last_synced_at"`
	LastSyncStatus SyncStatus    `json:"last_sync_status" gorm:"column:last_sync_status;default:pending"`
	LastSyncError  string        `json:"last_sync_error" gorm:"column:last_sync_error"`
}
