package main

import (
	"deploy-hub/config"
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
	router.Run() 
}