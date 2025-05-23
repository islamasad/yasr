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
	"yasr/internal/database"
	"yasr/pkg/models"
	"yasr/pkg/templates"
)

func main() {
	_ = godotenv.Load()

	// Handle command line arguments
	if len(os.Args) > 1 {
		commandHandled := handleCommands()
		if commandHandled {
			return // Hentikan eksekusi jika command valid
		}
	}

	// Jika tidak ada argumen, jalankan server
	db := setupDatabase()

	r := setupGinEngine()

	setupSessions(r)

	r.HTMLRender = templates.SetupTemplates()

	api.SetupRoutes(r, db)

	startServer(r)
}

func handleCommands() bool {
	cmd := os.Args[1]
	switch cmd {
	case "--reset-db":
		resetDatabase()
		return true
	case "--seed":
		seedDatabase()
		return true // Akan exit di dalam seedDatabase()
	default:
		log.Printf("Command tidak dikenali: %s", cmd)
		return false
	}
}

func resetDatabase() {
	dbName := os.Getenv("DB_NAME")

	err := database.DropDatabase(dbName)
	if err != nil {
		log.Printf("Gagal reset database: %v", err)
		return
	}

	err = database.CreateDatabase(dbName)
	if err != nil {
		log.Printf("Gagal create database: %v", err)
		return
	}

	db := setupDatabase()
	database.Seed(db)
	log.Println("Reset database berhasil")
}

func seedDatabase() {
	db := setupDatabase()

	// Tambahkan log awal
	log.Println("Memulai proses seeding...")

	database.Seed(db)

	// Verifikasi data
	var count int64
	db.Model(&models.Product{}).Count(&count)
	log.Printf("Seeding selesai. Total produk: %d", count)

	// Exit setelah selesai
	os.Exit(0)
}

func setupDatabase() *gorm.DB {
	dbName := os.Getenv("DB_NAME")

	// Step 1: Coba connect ke database target
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		dbName,
		os.Getenv("DB_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err == nil {
		return runMigrations(db)
	}

	// Step 2: Jika database tidak ada, buat database baru
	log.Println("Database tidak ditemukan, mencoba membuat database baru...")

	// Connect ke database default postgres
	adminDsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	adminDb, err := gorm.Open(postgres.Open(adminDsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal connect ke postgres: %v", err)
	}

	// Check apakah database sudah ada
	var exists bool
	adminDb.Raw(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)",
		dbName,
	).Scan(&exists)

	if !exists {
		// Create database
		err = adminDb.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
		if err != nil {
			log.Fatalf("Gagal membuat database: %v", err)
		}
		log.Printf("Database %s berhasil dibuat", dbName)
	}

	// Step 3: Connect ke database yang baru dibuat
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Gagal connect setelah membuat database: %v", err)
	}

	return runMigrations(db)
}

func runMigrations(db *gorm.DB) *gorm.DB {
	// Auto migrate models
	err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.OrderSession{},
		&models.OrderItem{},
	)
	if err != nil {
		log.Fatalf("Migrasi gagal: %v", err)
	}

	// Tambahkan extension UUID jika belum ada
	db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto")

	log.Println("Database siap digunakan")
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
