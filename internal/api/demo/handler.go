// internal/api/demo/handler.go
package demo

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"

	"yasr/pkg/models"
)

func DemoDashboardHandler(c *gin.Context) {
	// Ambil atau generate reference ID (contoh: dari database)
	referenceID := "MEJA-05" // Bisa diambil dari database

	orderURL := getAbsoluteURL(c, "/demo/order?ref="+referenceID)
	// Generate QR code

	c.HTML(http.StatusOK, "demo/dashboard", gin.H{

		"ReferenceID": referenceID,
		"OrderURL":    orderURL,
	})
}

func DemoOrderHandler(c *gin.Context) {

	// Ambil reference ID dari QR code
	referenceID := c.Query("ref")
	if referenceID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Reference ID tidak valid"})
		return
	}

	// Buat atau dapatkan session
	session := sessions.Default(c)
	log.Println("Session sebelum save:", session.Get("session_id")) // Debug
	sessionID := session.Get("session_id")
	if sessionID == nil {
		sessionID = uuid.New().String()
		session.Set("session_id", sessionID)
	}

	session.Set("reference_id", referenceID) // Simpan reference ID
	session.Set("created_at", time.Now().Unix())
	if err := session.Save(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan session",
		})
		log.Printf("Gagal save session: %v", err)
		return
	}

	c.HTML(http.StatusOK, "demo/order", gin.H{
		"ReferenceID": referenceID,
		"SessionID":   sessionID,
	})
}

func DemoCashierHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "demo/cashier", gin.H{
		"Title": "Kasir",
	})
}

func DemoQRHandler(c *gin.Context) {
	// Ambil atau generate reference ID (contoh: dari database)
	referenceID := "MEJA-05" // Bisa diambil dari database

	orderURL := getAbsoluteURL(c, "/demo/order?ref="+referenceID)

	png, err := qrcode.Encode(orderURL, qrcode.Medium, 256)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}

// Helper functions
func getAbsoluteURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, path)
}

func MenuHandler(c *gin.Context) {
	menu := getDemoMenu()

	type MenuItem struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Price    int    `json:"price"`
		Category string `json:"category"`
	}

	var response []MenuItem
	for _, m := range menu {
		response = append(response, MenuItem{
			ID:       m["id"].(int),
			Name:     m["name"].(string),
			Price:    m["price"].(int),
			Category: m["category"].(string),
		})
	}

	c.JSON(200, gin.H{"data": response})
}

// Contoh data menu
func getDemoMenu() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":       1,
			"name":     "Kopi Hitam",
			"price":    15000,
			"category": "Minuman",
		},
		{
			"id":       2,
			"name":     "Croissant",
			"price":    25000,
			"category": "Makanan",
		},
		{
			"id":       3,
			"name":     "Sandwich",
			"price":    30000,
			"category": "Makanan",
		},
		{
			"id":       4,
			"name":     "Cappuccino",
			"price":    20000,
			"category": "Minuman",
		},
	}
}

func SubmitOrderHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	session := sessions.Default(c)

	log.Println("Session di submit:", session.Get("session_id"))

	// Validasi session
	sessionID := session.Get("session_id")
	referenceID := session.Get("reference_id")
	createdAt := session.Get("created_at")

	log.Println("Session keys:", session.Get("session_id"), session.Get("reference_id"), session.Get("created_at"))

	if sessionID == nil || referenceID == nil || createdAt == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session tidak valid"})
		return
	}

	// Bind data dari form
	var orderItems []models.OrderItem
	if err := c.ShouldBindJSON(&orderItems); err != nil { // Ganti ke ShouldBindJSON
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hitung subtotal untuk setiap item
	for i := range orderItems {
		orderItems[i].Subtotal = orderItems[i].Price * orderItems[i].Quantity // Diubah dari Total ke Subtotal
		orderItems[i].SessionID = sessionID.(string)
	}

	// Hitung total order
	var total int
	for _, item := range orderItems {
		total += item.Subtotal // Menggunakan Subtotal untuk kalkulasi
	}

	// Simpan ke database menggunakan transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	orderSession := models.OrderSession{
		SessionID:   sessionID.(string),
		ReferenceID: referenceID.(string),
		CreatedAt:   createdAt.(time.Time),
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		Items:       orderItems,
		Total:       total,
	}
	if err := tx.Create(&orderSession).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan order"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":  "Order berhasil disimpan",
		"order_id": orderSession.ID,
	})
}

// Handler untuk update status
func UpdateOrderStatus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	orderID := c.Param("id")

	result := db.Model(&models.OrderSession{}).Where("id = ?", orderID).Update("status", input.Status)

	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

func GetOrdersHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var orders []models.OrderSession
	result := db.Preload("Items").Find(&orders)

	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat data order"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
