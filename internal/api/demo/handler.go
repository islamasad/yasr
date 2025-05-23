// internal/api/demo/handler.go
package demo

import (
	"fmt"
	"log"
	"net/http"
	"sync"
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
	db := c.MustGet("db").(*gorm.DB)

	var products []models.Product
	result := db.Find(&products)

	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil menu"})
		return
	}

	c.JSON(200, gin.H{"data": products})
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

// internal/api/demo/handler.go
func SubmitOrderHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	session := sessions.Default(c)

	// Validasi session
	sessionID := session.Get("session_id")
	referenceID := session.Get("reference_id")
	createdAt := session.Get("created_at")

	if sessionID == nil || referenceID == nil || createdAt == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session tidak valid"})
		return
	}

	// Konversi created_at ke Unix timestamp
	createdAtUnix, ok := createdAt.(int64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Format waktu tidak valid"})
		return
	}

	var requestItems []struct {
		ProductUUID string `json:"product_uuid"`
		Quantity    int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&requestItems); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi cart tidak kosong
	if len(requestItems) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Keranjang belanja kosong"})
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Hitung total dan siapkan order items
	var total int
	var orderItems []models.OrderItem

	for _, reqItem := range requestItems {
		// Parse UUID
		productUUID, err := uuid.Parse(reqItem.ProductUUID)
		if err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "UUID produk tidak valid"})
			return
		}

		// Dapatkan info produk
		var product models.Product
		if err := tx.Where("uuid = ?", productUUID).First(&product).Error; err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Produk tidak ditemukan"})
			return
		}

		// Hitung subtotal
		subtotal := product.Price * reqItem.Quantity
		total += subtotal

		orderItems = append(orderItems, models.OrderItem{
			ProductUUID: productUUID,
			Name:        product.Name,
			Price:       product.Price,
			Quantity:    reqItem.Quantity,
			Subtotal:    subtotal,
		})
	}

	// Buat order session
	orderSession := models.OrderSession{
		SessionID:   sessionID.(string),
		ReferenceID: referenceID.(string),
		CreatedAt:   time.Unix(createdAtUnix, 0),
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		Total:       total,
		Items:       orderItems,
	}

	if err := tx.Create(&orderSession).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan order"})
		return
	}

	tx.Commit()

	// Broadcast ke client
	BroadcastNewOrder(orderSession)

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

// internal/api/demo/handler.go
func OrderStreamHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	clientChan := make(chan models.OrderSession, 10)
	defer close(clientChan)

	registerClient(clientChan)
	defer unregisterClient(clientChan)

	go func() {
		<-c.Writer.CloseNotify()
		unregisterClient(clientChan)
	}()

	// Kirim data awal
	for _, order := range getInitialOrders(db) {
		sendOrderEvent(c, order)
		c.Writer.Flush()
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case order := <-clientChan:
			sendOrderEvent(c, order)
			c.Writer.Flush()
		case <-ticker.C:
			c.SSEvent("ping", nil)
			c.Writer.Flush()
		}
	}
}

func sendOrderEvent(c *gin.Context, order models.OrderSession) {
	data := map[string]interface{}{
		"id":         order.ID,
		"created_at": order.CreatedAt.Format(time.RFC3339),
		"items":      order.Items,
		"total":      order.Total,
		"status":     order.Status,
	}
	c.SSEvent("message", data)
}

// Fungsi helper untuk broadcast
var clients = make(map[chan models.OrderSession]bool)
var mutex sync.Mutex

func registerClient(clientChan chan models.OrderSession) {
	mutex.Lock()
	defer mutex.Unlock()
	clients[clientChan] = true
}

func unregisterClient(clientChan chan models.OrderSession) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(clients, clientChan)
}

// Fungsi untuk mengirim update order baru ke semua client
func BroadcastNewOrder(order models.OrderSession) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		select {
		case client <- order: // Non-blocking send
		default:
			log.Println("Client channel full, skipping")
		}
	}
}

func getInitialOrders(db *gorm.DB) []models.OrderSession {
	var orders []models.OrderSession
	result := db.Preload("Items").
		Order("created_at DESC").
		Limit(50). // Batasi 50 order terakhir
		Find(&orders)

	if result.Error != nil {
		log.Printf("Error fetching initial orders: %v", result.Error)
		return []models.OrderSession{}
	}
	return orders
}
