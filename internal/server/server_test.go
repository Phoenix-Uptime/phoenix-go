package server

import (
	"bytes"
	"context"
	stdsql "database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
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

func useTestDatabase(t *testing.T) func() {
	t.Helper()

	db, err := stdsql.Open("sqlite", filepath.Join(t.TempDir(), "phoenix-test.db"))
	if err != nil {
		t.Fatal(err)
	}
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
