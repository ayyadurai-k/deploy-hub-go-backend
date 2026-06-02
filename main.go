package main

import (
	"deploy-hub/config"
	"deploy-hub/internal/accounts"
	"net/http"

	"github.com/gin-gonic/gin"
)


func init(){
	config.LoadEnvVariables()
	config.ConnectDB()
}
func main()  {
	router := gin.Default()

	router.GET("/healthz",func (c *gin.Context)  {
		c.JSON(http.StatusOK,gin.H{
			"health":"ok",
		})
	})

	accounts.RegisterRoutes(router.Group("/auth"))




	router.Run() 
}