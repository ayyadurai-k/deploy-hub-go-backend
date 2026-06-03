package service

import (
	"time"

	"gorm.io/gorm/clause"

	"deploy-hub/config"
)

type BlacklistedToken struct {
	ID        uint      `gorm:"primarykey"`
	JTI       string    `gorm:"column:jti;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time
}

func BlacklistJTI(jti string, expiresAt time.Time) (bool, error) {
	result := config.DB.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&BlacklistedToken{JTI: jti, ExpiresAt: expiresAt})
	return result.RowsAffected > 0, result.Error
}

func IsBlacklisted(jti string) bool {
	var count int64
	config.DB.Model(&BlacklistedToken{}).Where("jti = ?", jti).Count(&count)
	return count > 0
}
