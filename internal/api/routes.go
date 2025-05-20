package api

import (
	"net/http"
	"yasr/internal/api/auth"
	"yasr/internal/api/demo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Public routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{"Title": "Home"})
	})

	// Auth routes
	auth.RegisterAuthRoutes(r.Group(""), db)

	// Demo routes
	demo.RegisterDemoRoutes(r.Group(""), db)

	// Authenticated routes
	authGroup := r.Group("/dashboard", auth.AuthRequired())
	{
		authGroup.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard", gin.H{"Title": "Dashboard"})
		})
	}
}
