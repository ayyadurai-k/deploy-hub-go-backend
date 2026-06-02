package oauth

import "github.com/gin-gonic/gin"


func RegisterRoutes(rg *gin.RouterGroup){
	rg.POST("/google/start")
	rg.POST("/google/callback")
	rg.POST("/github/start")
	rg.POST("/github/callback")
}