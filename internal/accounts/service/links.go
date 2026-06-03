package service

import "deploy-hub/config"

func ProviderLinks(userID uint) (hasGoogle, hasGithub bool) {
	var google, github int64
	config.DB.Table("google_profiles").Where("user_id = ?", userID).Count(&google)
	config.DB.Table("git_hub_profiles").Where("user_id = ?", userID).Count(&github)
	return google > 0, github > 0
}
