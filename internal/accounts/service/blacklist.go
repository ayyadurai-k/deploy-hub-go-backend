package service

import (
	"time"

	"deploy-hub/config"
)

type BlacklistedToken struct {
	ID        uint      `gorm:"primarykey"`
	JTI       string    `gorm:"column:jti;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time
}

func BlacklistJTI(jti string, expiresAt time.Time) error {
	return config.DB.Create(&BlacklistedToken{JTI: jti, ExpiresAt: expiresAt}).Error
}

func IsBlacklisted(jti string) bool {
	var count int64
	config.DB.Model(&BlacklistedToken{}).Where("jti = ?", jti).Count(&count)
	return count > 0
}
