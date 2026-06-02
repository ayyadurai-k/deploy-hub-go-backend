package accounts

import "github.com/gin-gonic/gin"


func RegisterRoutes(rg *gin.RouterGroup){
	rg.POST("/me")
	rg.POST("/refresh")
	rg.POST("/logout")
}