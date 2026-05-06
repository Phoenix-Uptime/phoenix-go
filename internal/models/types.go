package models

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

type Filters struct {
	Contains    string `json:"contains,omitempty" swagger:"example=Success"`
	NotContains string `json:"not_contains,omitempty" swagger:"example=Error"`
}

type MessageType string

const (
	MessageIssue       MessageType = "issue"
	MessageInvestigate MessageType = "investigate"
	MessageResolved    MessageType = "resolved"
)

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
