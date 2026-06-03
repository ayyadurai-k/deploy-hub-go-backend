package oauth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/google/start", GoogleStart)
	rg.GET("/google/callback", GoogleCallback)
	rg.GET("/github/start", GitHubStart)
	rg.GET("/github/callback", GithubCallback)
}
