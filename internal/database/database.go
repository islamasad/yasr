package database

import (
	"fmt"
	"log"
	"os"

	"yasr/pkg/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DropDatabase(dbName string) error {
	// Connect ke database default
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("gagal connect ke postgres: %v", err)
	}

	// Terminate active connections
	db.Exec(fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
		AND pid <> pg_backend_pid()`, dbName))

	// Drop database
	err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).Error
	if err != nil {
		return fmt.Errorf("gagal drop database: %v", err)
	}

	log.Printf("Database %s berhasil dihapus", dbName)
	return nil
}

func CreateDatabase(dbName string) error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Create database
	err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
	if err != nil {
		return fmt.Errorf("gagal create database: %v", err)
	}

	log.Printf("Database %s berhasil dibuat", dbName)
	return nil
}

func Seed(db *gorm.DB) {
	seedProducts(db)
}

func seedProducts(db *gorm.DB) {
	products := []models.Product{
		{
			Name:        "Kopi Hitam",
			Price:       15000,
			Category:    "Minuman",
			Description: "Kopi arabika pilihan",
		},
		{
			Name:        "Croissant",
			Price:       25000,
			Category:    "Makanan",
			Description: "Croissant dengan butter premium",
		},
		{
			Name:        "Sandwich",
			Price:       30000,
			Category:    "Makanan",
			Description: "Roti gandum dengan isian segar",
		},
		{
			Name:        "Cappuccino",
			Price:       20000,
			Category:    "Minuman",
			Description: "Cappuccino dengan foam lembut",
		},
	}

	for _, p := range products {
		result := db.Where(models.Product{Name: p.Name}).FirstOrCreate(&p)
		if result.Error != nil {
			log.Printf("❌ Gagal seeding produk %s: %v", p.Name, result.Error)
		} else if result.RowsAffected == 1 {
			log.Printf("✅ Berhasil membuat produk: %s", p.Name)
		} else {
			log.Printf("⏩ Produk %s sudah ada", p.Name)
		}
	}
}
