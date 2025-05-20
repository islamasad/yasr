package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterAuthRoutes(r *gin.RouterGroup, db *gorm.DB) {
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "auth/login", gin.H{"Title": "Login"})
		})
		authGroup.GET("/register", func(c *gin.Context) {
			c.HTML(http.StatusOK, "auth/register", gin.H{"Title": "Register"})
		})
		authGroup.POST("/login", func(c *gin.Context) { LoginHandler(c, db) })
		authGroup.POST("/register", func(c *gin.Context) { RegisterHandler(c, db) })
		authGroup.GET("/logout", LogoutHandler)
	}
}
