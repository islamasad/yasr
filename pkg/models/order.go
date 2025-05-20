// 1. Update models/order.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderSession struct {
	gorm.Model
	SessionID   string `gorm:"uniqueIndex;size:255"`
	ReferenceID string `gorm:"size:255"` // Bisa berupa meja, kode unik, dll
	Status      string `gorm:"size:50;default:'pending'"`
	CreatedAt   time.Time
	ExpiredAt   time.Time
	Items       []OrderItem `gorm:"foreignKey:SessionID;references:SessionID"`
	Total       int         `gorm:"default:0"`
}

type OrderItem struct {
	gorm.Model
	SessionID string `gorm:"size:255"`
	Name      string
	Quantity  int
	Price     int
	Subtotal  int
}
