// models/product.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex" json:"uuid"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Price       int            `gorm:"not null" json:"price"`
	Category    string         `gorm:"size:100;not null" json:"category"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
