package service

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func IssueJWTPair(userID uint) (IssuedPair, error) {
	access, err := signToken(userID, "access", accessLifetime())
	if err != nil {
		return IssuedPair{}, err
	}
	refresh, err := signToken(userID, "refresh", refreshLifetime())
	if err != nil {
		return IssuedPair{}, err
	}
	return IssuedPair{Access: access, Refresh: refresh}, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) { return secret(), nil },
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

func signToken(userID uint, tokenType string, lifetime time.Duration) (string, error) {
	jti, err := newJTI()
	if err != nil {
		return "", err
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
		},
	})
	return token.SignedString(secret())
}

func newJTI() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func secret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
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

func RefreshCookieName() string {
	return refreshCookieName()
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
	return "/api/v1/auth/"
}

func refreshCookieSameSite() http.SameSite {
	switch strings.ToLower(os.Getenv("REFRESH_COOKIE_SAMESITE")) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
