// pkg/models/user.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User merepresentasikan tabel users.
// - ID: SERIAL primary key, untuk penggunaan internal.
// - UUID: eksternal identifier yang aman.
// - Email, Password, Name, Role sesuai skema.
// - CreatedAt, UpdatedAt otomatis di-manage oleh GORM.
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex" json:"uuid"`
	Email     string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Role      string         `gorm:"size:50;not null;default:user" json:"role"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
