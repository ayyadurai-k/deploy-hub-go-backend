package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"deploy-hub/config"
	"deploy-hub/internal/accounts"
	accountsvc "deploy-hub/internal/accounts/service"
	"deploy-hub/internal/httpx"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found || token == "" {
			httpx.Error(c, http.StatusUnauthorized, "not_authenticated", "missing bearer token")
			c.Abort()
			return
		}

		claims, err := accountsvc.ParseToken(token)
		if err != nil || claims.TokenType != "access" {
			httpx.Error(c, http.StatusUnauthorized, "not_authenticated", "invalid token")
			c.Abort()
			return
		}

		var user accounts.User
		if err := config.DB.First(&user, claims.UserID).Error; err != nil {
			httpx.Error(c, http.StatusUnauthorized, "not_authenticated", "user not found")
			c.Abort()
			return
		}

		c.Set(accounts.CtxUserKey, user)
		c.Next()
	}
}
