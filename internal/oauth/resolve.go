package oauth

import (
	"errors"

	"gorm.io/gorm"

	"deploy-hub/config"
	"deploy-hub/internal/accounts"
	"deploy-hub/internal/oauth/service"
)

var ErrNoVerifiedEmail = errors.New("github account has no verified primary email")

type ResolutionResult struct {
	User           accounts.User
	ProfileCreated bool
	UserCreated    bool
}

func ResolveGoogle(identity service.GoogleIdentity, tokens service.GoogleTokens) (ResolutionResult, error) {
	var result ResolutionResult

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var existing GoogleProfile
		err := tx.Preload("User").Where("google_sub = ?", identity.Sub).First(&existing).Error

		switch {
		case err == nil:
			if identity.Email != "" {
				existing.Email = identity.Email
			}
			if identity.Picture != "" {
				existing.PictureURL = identity.Picture
			}
			if err := existing.SetAccessToken(tokens.AccessToken); err != nil {
				return err
			}
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			if identity.Email != "" && existing.User.Email != identity.Email {
				if err := tx.Model(&existing.User).Update("email", identity.Email).Error; err != nil {
					return err
				}
			}
			result = ResolutionResult{User: existing.User, ProfileCreated: false, UserCreated: false}
			return nil

		case errors.Is(err, gorm.ErrRecordNotFound):
			user := accounts.User{
				Email:       identity.Email,
				DisplayName: identity.Name,
				IsActive:    true,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			profile := GoogleProfile{
				UserID:     user.ID,
				GoogleSub:  identity.Sub,
				Email:      identity.Email,
				PictureURL: identity.Picture,
			}
			if err := profile.SetAccessToken(tokens.AccessToken); err != nil {
				return err
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
			result = ResolutionResult{User: user, ProfileCreated: true, UserCreated: true}
			return nil

		default:
			return err
		}
	})

	return result, err
}

func ResolveGitHub(identity service.GitHubIdentity, tokens service.GitHubTokens) (ResolutionResult, error) {
	var result ResolutionResult

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var existing GitHubProfile
		err := tx.Preload("User").Where("github_user_id = ?", identity.UserID).First(&existing).Error

		switch {
		case err == nil:
			if err := existing.SetAccessToken(tokens.AccessToken); err != nil {
				return err
			}
			if identity.Login != "" {
				existing.GitHubLogin = identity.Login
			}
			if identity.AvatarURL != "" {
				existing.AvatarURL = identity.AvatarURL
			}
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			if identity.Email != "" && existing.User.Email != identity.Email {
				if err := tx.Model(&existing.User).Update("email", identity.Email).Error; err != nil {
					return err
				}
			}
			result = ResolutionResult{User: existing.User, ProfileCreated: false, UserCreated: false}
			return nil

		case errors.Is(err, gorm.ErrRecordNotFound):
			if identity.Email == "" {
				return ErrNoVerifiedEmail
			}

			displayName := identity.Name
			if displayName == "" {
				displayName = identity.Login
			}
			user := accounts.User{
				Email:       identity.Email,
				DisplayName: displayName,
				IsActive:    true,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			profile := GitHubProfile{
				UserID:       user.ID,
				GitHubUserID: identity.UserID,
				GitHubLogin:  identity.Login,
				AvatarURL:    identity.AvatarURL,
			}
			if err := profile.SetAccessToken(tokens.AccessToken); err != nil {
				return err
			}
			if err := tx.Create(&profile).Error; err != nil {
				return err
			}
			result = ResolutionResult{User: user, ProfileCreated: true, UserCreated: true}
			return nil

		default:
			return err
		}
	})

	return result, err
}
