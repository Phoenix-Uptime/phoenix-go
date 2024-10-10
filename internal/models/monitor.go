package models

// Monitor struct to hold monitor details
type Monitor struct {
	Model
	UserID     uint             `json:"user_id" gorm:"index:idx_user_id"`
	Name       string           `gorm:"not null" json:"name" swagger:"example=My URL Monitor"`
	URL        string           `gorm:"not null" json:"url" swagger:"example=https://example.com"`
	Interval   int              `gorm:"default:60" json:"interval" swagger:"example=60"` // Interval in seconds
	Status     Status           `gorm:"type:enum('up','down','paused','unknown');default:'unknown'" json:"status" swagger:"example=up"`
	Type       MonitorType      `gorm:"not null" json:"type" swagger:"example=url"`
	Method     *string          `json:"method,omitempty" swagger:"example=GET"`                    // For URL monitors
	Filters    *Filters         `gorm:"embedded;embeddedPrefix:filters_" json:"filters,omitempty"` // Optional filters
	Tags       []Tag            `gorm:"many2many:monitor_tags;joinForeignKey:MonitorID;joinReferences:TagID" json:"tags"`
	SMTPConfig *SMTPMonitor     `gorm:"embedded;embeddedPrefix:smtp_" json:"smtp_config,omitempty"` // Optional SMTP settings
	History    []MonitorHistory `gorm:"foreignKey:MonitorID" json:"history"`
	Retry      int              `json:"retry" gorm:"default:3" swagger:"example=3"`         // Number of retries
	RetryAfter int              `json:"retry_after" gorm:"default:30" swagger:"example=30"` // Seconds between retries
}

// MonitorHistory struct to hold history related to each monitor
type MonitorHistory struct {
	Model
	MonitorID    uint   `json:"monitor_id" gorm:"index:idx_monitor_id"`
	Status       string `json:"status" swagger:"example=up"`
	ResponseTime int    `json:"response_time" swagger:"example=200"`
	ErrorMessage string `json:"error_message,omitempty"`
}
