package service

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"deploy-hub/internal/accounts"
)

type IssuedPair struct {
	Access  string
	Refresh string
}

type Claims struct {
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func IssueJWTPair(user accounts.User) (IssuedPair, error) {
	access, err := signToken(user.ID, "access", accessLifetime())
	if err != nil {
		return IssuedPair{}, err
	}
	refresh, err := signToken(user.ID, "refresh", refreshLifetime())
	if err != nil {
		return IssuedPair{}, err
	}
	return IssuedPair{Access: access, Refresh: refresh}, nil
}

func signToken(userID uint, tokenType string, lifetime time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
		},
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func SetRefreshCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName(),
		Value:    refreshToken,
		Path:     refreshCookiePath(),
		Domain:   os.Getenv("REFRESH_COOKIE_DOMAIN"),
		MaxAge:   int(refreshLifetime().Seconds()),
		Secure:   os.Getenv("REFRESH_COOKIE_SECURE") == "true",
		HttpOnly: true,
		SameSite: refreshCookieSameSite(),
	})
}

func ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName(),
		Value:    "",
		Path:     refreshCookiePath(),
		Domain:   os.Getenv("REFRESH_COOKIE_DOMAIN"),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: refreshCookieSameSite(),
	})
}

func accessLifetime() time.Duration {
	return envDuration("ACCESS_TOKEN_LIFETIME", 15*time.Minute)
}

func refreshLifetime() time.Duration {
	return envDuration("REFRESH_TOKEN_LIFETIME", 7*24*time.Hour)
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func refreshCookieName() string {
	if value := os.Getenv("REFRESH_COOKIE_NAME"); value != "" {
		return value
	}
	return "refresh_token"
}

func refreshCookiePath() string {
	if value := os.Getenv("REFRESH_COOKIE_PATH"); value != "" {
		return value
	}
	return "/"
}

func refreshCookieSameSite() http.SameSite {
	switch strings.ToLower(os.Getenv("REFRESH_COOKIE_SAMESITE")) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
