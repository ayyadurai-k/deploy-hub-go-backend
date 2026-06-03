package repositories

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	rg.GET("", requireAuth, List)
	rg.POST("/sync", requireAuth, Sync)
}
