package server

import (
	"net/http"
	"testing"

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
