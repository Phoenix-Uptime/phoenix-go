package models

// Monitor Status Type
type Status string

const (
	StatusUp      Status = "up"
	StatusDown    Status = "down"
	StatusPaused  Status = "paused"
	StatusUnknown Status = "unknown"
)

// Monitor Type Type
type MonitorType string

const (
	TypeURL  MonitorType = "url"
	TypePing MonitorType = "ping"
	TypeSMTP MonitorType = "smtp"
)
