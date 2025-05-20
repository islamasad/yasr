package auth

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"yasr/pkg/models"
)

// RegisterHandler menerima JSON { email, password, name }
func RegisterHandler(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hash, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	user := models.User{
		Email:    payload.Email,
		Password: string(hash),
		Name:     payload.Name,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already used"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"uuid": user.UUID})
}

// LoginHandler menerima JSON { email, password }
func LoginHandler(c *gin.Context, db *gorm.DB) {
	var payload struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Set session
	sess := sessions.Default(c)
	sess.Set("user_uuid", user.UUID.String())
	sess.Save()

	c.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

// LogoutHandler clear session
func LogoutHandler(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	sess.Save()
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// ProfileHandler contoh route ter‐proteksi
func ProfileHandler(c *gin.Context, db *gorm.DB) {
	sess := sessions.Default(c)
	uuid := sess.Get("user_uuid")
	if uuid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var user models.User
	if err := db.Where("uuid = ?", uuid.(string)).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	// Kirim data profil (tanpa password)
	c.JSON(http.StatusOK, gin.H{
		"uuid":  user.UUID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}
