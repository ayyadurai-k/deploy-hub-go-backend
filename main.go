package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"deploy-hub/config"
	"deploy-hub/internal/accounts"
	accountsvc "deploy-hub/internal/accounts/service"
	"deploy-hub/internal/middleware"
	"deploy-hub/internal/oauth"
	"deploy-hub/internal/repositories"
)

func init() {
	config.LoadEnvVariables()
	config.ConnectDB()
	config.DB.AutoMigrate(
		&accounts.User{},
		&oauth.GoogleProfile{},
		&oauth.GitHubProfile{},
		&repositories.Repository{},
		&accountsvc.BlacklistedToken{},
	)
}

func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.CORSAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	v1 := router.Group("/api/v1")

	v1.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	v1.GET("/readyz", func(c *gin.Context) {
		sqlDB, err := config.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	requireAuth := middleware.RequireAuth()

	accounts.RegisterRoutes(v1.Group("/auth"), requireAuth)
	oauth.RegisterRoutes(v1.Group("/oauth"))
	repositories.RegisterRoutes(v1.Group("/repositories"), requireAuth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router.Run(":" + port)
}
