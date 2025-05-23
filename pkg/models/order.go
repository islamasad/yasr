// 1. Update models/order.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderSession struct {
	gorm.Model
	SessionID   string `gorm:"uniqueIndex;size:255"`
	ReferenceID string `gorm:"size:255"` // Bisa berupa meja, kode unik, dll
	Status      string `gorm:"size:50;default:'pending'"`
	CreatedAt   time.Time
	ExpiredAt   time.Time
	Items       []OrderItem `gorm:"foreignKey:OrderSessionID;references:ID"`
	Total       int         `gorm:"default:0"`
}

type OrderItem struct {
	gorm.Model
	OrderSessionID uint      `json:"-"`
	ProductUUID    uuid.UUID `gorm:"type:uuid;not null" json:"product_uuid"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	Quantity       int       `gorm:"not null" json:"quantity"`
	Price          int       `gorm:"not null" json:"price"`
	Subtotal       int       `gorm:"not null" json:"subtotal"`
}
