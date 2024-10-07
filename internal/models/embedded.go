package models

type SMTPMonitor struct {
	SMTPServer string `json:"smtp_server" swagger:"example=smtp.example.com"`
	SMTPPort   int    `json:"smtp_port" swagger:"example=587"`
	Username   string `json:"username" swagger:"example=user@example.com"`
	Password   string `json:"password" swagger:"example=supersecret"`
}

type Filters struct {
	Contains    string `json:"contains,omitempty" swagger:"example=Success"`
	NotContains string `json:"not_contains,omitempty" swagger:"example=Error"`
}
