package oauth

import (

	"github.com/gin-gonic/gin"

)


func RegisterRoutes(rg *gin.RouterGroup){
	rg.POST("/google/start",GoogleStart)
	rg.POST("/google/callback")
	rg.POST("/github/start",GitHubStart)
	rg.POST("/github/callback")
}