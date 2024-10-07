package models

import (
	"time"

	"gorm.io/gorm"
)

// Model is a base model that includes ID, CreatedAt, UpdatedAt, and DeletedAt fields.
type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
