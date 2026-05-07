package server

import (
	"bytes"
	"context"
	stdsql "database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/api"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
)

func TestHealthRoute(t *testing.T) {
	app := New()

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAccountSettingsRoutesRequireAuth(t *testing.T) {
	app := New()

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/account/settings"},
		{method: http.MethodPost, path: "/account/settings/smtp"},
		{method: http.MethodPost, path: "/account/settings/telegram"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}

func TestSignupAndLoginWithEnt(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()

	signupBody := []byte(`{"username":"exampleuser","email":"user@example.com","password":"examplepassword"}`)
	signupReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(signupBody))
	if err != nil {
		t.Fatal(err)
	}
	signupReq.Header.Set("Content-Type", "application/json")

	signupResp, err := app.Test(signupReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if signupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d", http.StatusCreated, signupResp.StatusCode)
	}

	loginBody := []byte(`{"username":"exampleuser","password":"examplepassword"}`)
	loginReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected login status %d, got %d", http.StatusOK, loginResp.StatusCode)
	}
	defer loginResp.Body.Close()

	var body api.LoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.ApiKey == "" {
		t.Fatal("expected login response to include api key")
	}
}

func TestAccountSettingsUseNotificationChannelRows(t *testing.T) {
	cleanup := useTestDatabase(t)
	defer cleanup()

	app := New()

	signupBody := []byte(`{"username":"settingsuser","email":"settings@example.com","password":"examplepassword"}`)
	signupReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewReader(signupBody))
	if err != nil {
		t.Fatal(err)
	}
	signupReq.Header.Set("Content-Type", "application/json")

	signupResp, err := app.Test(signupReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if signupResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d", http.StatusCreated, signupResp.StatusCode)
	}

	loginBody := []byte(`{"username":"settingsuser","password":"examplepassword"}`)
	loginReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := app.Test(loginReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()

	var login api.LoginResponse
	if err := json.NewDecoder(loginResp.Body).Decode(&login); err != nil {
		t.Fatal(err)
	}
	if login.ApiKey == "" {
		t.Fatal("expected login response to include api key")
	}

	smtpBody := []byte(`{"smtp_server":"smtp.example.com","smtp_port":587,"from_address":"noreply@example.com","username":"mailer@example.com","password":"supersecret","use_tls":true}`)
	smtpReq, err := http.NewRequest(http.MethodPost, "/account/settings/smtp", bytes.NewReader(smtpBody))
	if err != nil {
		t.Fatal(err)
	}
	smtpReq.Header.Set("Content-Type", "application/json")
	smtpReq.Header.Set("x-api-key", login.ApiKey)

	smtpResp, err := app.Test(smtpReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if smtpResp.StatusCode != http.StatusOK {
		t.Fatalf("expected SMTP settings status %d, got %d", http.StatusOK, smtpResp.StatusCode)
	}

	telegramBody := []byte(`{"bot_token":"123456789:ABCdefGHIjklMNOpqrSTUvwxyz"}`)
	telegramReq, err := http.NewRequest(http.MethodPost, "/account/settings/telegram", bytes.NewReader(telegramBody))
	if err != nil {
		t.Fatal(err)
	}
	telegramReq.Header.Set("Content-Type", "application/json")
	telegramReq.Header.Set("x-api-key", login.ApiKey)

	telegramResp, err := app.Test(telegramReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if telegramResp.StatusCode != http.StatusOK {
		t.Fatalf("expected Telegram settings status %d, got %d", http.StatusOK, telegramResp.StatusCode)
	}

	settingsReq, err := http.NewRequest(http.MethodGet, "/account/settings", nil)
	if err != nil {
		t.Fatal(err)
	}
	settingsReq.Header.Set("x-api-key", login.ApiKey)

	settingsResp, err := app.Test(settingsReq, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	if settingsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status %d, got %d", http.StatusOK, settingsResp.StatusCode)
	}
	defer settingsResp.Body.Close()

	var settings api.SettingsResponse
	if err := json.NewDecoder(settingsResp.Body).Decode(&settings); err != nil {
		t.Fatal(err)
	}
	if settings.SMTPSettings == nil || settings.SMTPSettings.SMTPServer != "smtp.example.com" {
		t.Fatal("expected settings response to include SMTP settings")
	}
	if settings.TelegramBot == nil || settings.TelegramBot.BotToken == "" {
		t.Fatal("expected settings response to include Telegram bot settings")
	}

	channelCount, err := database.Client.NotificationChannel.Query().
		Where(entnotificationchannel.TypeIn(entnotificationchannel.TypeSMTP, entnotificationchannel.TypeTelegram)).
		Count(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if channelCount != 2 {
		t.Fatalf("expected 2 notification channel rows, got %d", channelCount)
	}
}

func useTestDatabase(t *testing.T) func() {
	t.Helper()

	db, err := stdsql.Open("sqlite", "file:phoenix_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatal(err)
	}

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatal(err)
	}

	previous := database.Client
	database.Client = client
	return func() {
		database.Client = previous
		if err := client.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
