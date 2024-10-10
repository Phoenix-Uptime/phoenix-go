package models

import (
	"time"

	"gorm.io/gorm"
)

// Base model with standard fields
type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	Model
	Username     string        `gorm:"unique;not null;index:idx_username" json:"username" swagger:"example=exampleuser"`
	Email        string        `gorm:"unique;not null;index:idx_email" json:"email" swagger:"example=user@example.com"`
	Password     string        `gorm:"not null" json:"-"` // Password should not be returned in JSON
	ApiKey       string        `gorm:"unique;not null;index:idx_api_key" json:"api_key" swagger:"example=your_api_key"`
	Monitors     []Monitor     `gorm:"foreignKey:UserID" json:"monitors"`
	SMTPSettings *SMTPSettings `gorm:"embedded;embeddedPrefix:smtp_" json:"smtp_settings,omitempty"`    // Optional SMTP settings
	TelegramBot  *TelegramBot  `gorm:"embedded;embeddedPrefix=telegram_" json:"telegram_bot,omitempty"` // Optional Telegram bot settings
}

type SMTPSettings struct {
	SMTPServer  string `json:"smtp_server" swagger:"example=smtp.example.com"`
	SMTPPort    int    `json:"smtp_port" swagger:"example=587"`
	FromAddress string `json:"from_address" swagger:"example=noreply@example.com"`
	Username    string `json:"username" swagger:"example=user@example.com"`
	Password    string `json:"password" swagger:"example=supersecret"`
	UseTLS      bool   `json:"use_tls" swagger:"example=true"`
}

type TelegramBot struct {
	BotToken string `json:"bot_token" swagger:"example=123456789:ABCdefGHIjklMNOpqrSTUvwxyz"`
}

type Monitor struct {
	Model
	UserID     uint             `json:"user_id" gorm:"index:idx_user_id"`
	Name       string           `gorm:"not null" json:"name" swagger:"example=My URL Monitor"`
	URL        string           `gorm:"not null" json:"url" swagger:"example=https://example.com"`
	Interval   int              `gorm:"default:60" json:"interval" swagger:"example=60"` // Interval in seconds
	Status     Status           `gorm:"default:unknown" json:"status" swagger:"example=up"`
	Type       MonitorType      `gorm:"not null" json:"type" swagger:"example=url"`
	Method     *string          `json:"method,omitempty" swagger:"example=GET"`                    // For URL monitors
	Filters    *Filters         `gorm:"embedded;embeddedPrefix:filters_" json:"filters,omitempty"` // Optional filters
	Tags       []Tag            `gorm:"many2many:monitor_tags;joinForeignKey:MonitorID;joinReferences:TagID" json:"tags"`
	History    []MonitorHistory `gorm:"foreignKey:MonitorID" json:"history"`
	Retry      int              `json:"retry" gorm:"default:3" swagger:"example=3"`                          // Number of retries
	RetryAfter int              `json:"retry_after" gorm:"default:30" swagger:"example=30"`                  // Seconds between retries
	AlertTypes AlertTypes       `json:"alert_types" gorm:"type:json" swagger:"example=['smtp', 'telegram']"` // Array of alert types
}

type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusPaused  Status = "paused"
	StatusUnknown Status = "unknown"
)

type MonitorType string

const (
	TypeURL  MonitorType = "url"
	TypePing MonitorType = "ping"
	TypeSMTP MonitorType = "smtp"
)

type MonitorHistory struct {
	Model
	MonitorID    uint   `json:"monitor_id" gorm:"index:idx_monitor_id"`
	Status       string `json:"status" swagger:"example=up"`
	ResponseTime int    `json:"response_time" swagger:"example=200"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type Filters struct {
	Contains    string `json:"contains,omitempty" swagger:"example=Success"`
	NotContains string `json:"not_contains,omitempty" swagger:"example=Error"`
}

type Tag struct {
	Model
	Name        string    `gorm:"unique;not null" json:"name" swagger:"example=important"`
	Description string    `json:"description,omitempty" swagger:"example=Critical tag"`
	Color       string    `json:"color,omitempty" swagger:"example=#FF0000"`
	Monitors    []Monitor `gorm:"many2many:monitor_tags;" json:"monitors"`
}
