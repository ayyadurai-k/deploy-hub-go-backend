package httpx

import "github.com/gin-gonic/gin"

// Error writes the standard error envelope: {"error": {"code", "message"}}.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
