package accounts

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	rg.GET("/me", requireAuth, Me)
	rg.POST("/refresh", Refresh)
	rg.POST("/logout", Logout)
}
