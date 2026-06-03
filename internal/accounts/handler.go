package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"

	accountsvc "deploy-hub/internal/accounts/service"
	"deploy-hub/internal/httpx"
)

func currentUser(c *gin.Context) User {
	return c.MustGet(CtxUserKey).(User)
}

func Me(c *gin.Context) {
	c.JSON(http.StatusOK, ToUserResponse(currentUser(c)))
}

func Refresh(c *gin.Context) {
	raw, err := c.Cookie(accountsvc.RefreshCookieName())
	if err != nil || raw == "" {
		httpx.Error(c, http.StatusUnauthorized, "no_refresh_cookie", "refresh cookie missing")
		return
	}

	claims, err := accountsvc.ParseToken(raw)
	if err != nil || claims.TokenType != "refresh" {
		httpx.Error(c, http.StatusUnauthorized, "invalid_refresh", "invalid refresh token")
		return
	}
	if accountsvc.IsBlacklisted(claims.ID) {
		httpx.Error(c, http.StatusUnauthorized, "invalid_refresh", "refresh token revoked")
		return
	}

	// Rotate: blacklist the old token, then mint a fresh pair.
	if err := accountsvc.BlacklistJTI(claims.ID, claims.ExpiresAt.Time); err != nil {
		httpx.Error(c, http.StatusInternalServerError, "server_error", "could not rotate token")
		return
	}

	pair, err := accountsvc.IssueJWTPair(claims.UserID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "server_error", "could not issue tokens")
		return
	}

	accountsvc.SetRefreshCookie(c.Writer, pair.Refresh)
	c.JSON(http.StatusOK, gin.H{"access": pair.Access})
}

func Logout(c *gin.Context) {
	raw, err := c.Cookie(accountsvc.RefreshCookieName())
	if err == nil && raw != "" {
		if claims, parseErr := accountsvc.ParseToken(raw); parseErr == nil {
			accountsvc.BlacklistJTI(claims.ID, claims.ExpiresAt.Time)
		}
	}
	accountsvc.ClearRefreshCookie(c.Writer)
	c.Status(http.StatusNoContent)
}
