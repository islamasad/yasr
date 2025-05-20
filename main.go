package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"yasr/internal/api"
	"yasr/pkg/models"
	"yasr/pkg/templates"
)

func main() {
	_ = godotenv.Load()

	// Database setup
	db := setupDatabase()

	// Gin setup
	r := setupGinEngine()

	// Session middleware
	setupSessions(r)

	// Templates
	r.HTMLRender = templates.SetupTemplates()

	// Setup routes
	api.SetupRoutes(r, db)

	// Start server
	startServer(r)
}

func setupDatabase() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.OrderSession{},
		&models.OrderItem{}); err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	return db
}

func setupGinEngine() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // Ganti dengan origin frontend
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Static files
	r.Static("/dist", "./dist")
	r.Static("/src", "./src")

	return r
}

func setupSessions(r *gin.Engine) {
	gob.Register(time.Time{})
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET must be set in .env")
	}

	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 hari
		HttpOnly: true,
		Secure:   false, // false untuk development
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("mysession", store))
}

func startServer(r *gin.Engine) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s...\n", port)
	log.Fatal(r.Run(":" + port))
}
