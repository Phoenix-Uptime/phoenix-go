package api_keys

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entapikey "github.com/Phoenix-Uptime/phoenix-go/ent/apikey"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type APIKeyResponse struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Name       string     `json:"name"`
	KeyPreview string     `json:"key_preview"`
	IsActive   bool       `json:"is_active"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// @Summary Get API Key
// @Description Returns one API key owned by the authenticated user without exposing the secret.
// @Tags API Keys
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "API Key ID"
// @Success 200 {object} APIKeyResponse "API key"
// @Failure 400 {object} routes.ErrorResponse "invalid API key id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "API key not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /api-keys/{id} [get]
func GetAPIKey(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := apiKeyID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid API key id",
		})
	}

	key, err := userAPIKey(c, user.ID, id)
	if err != nil {
		return apiKeyLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(apiKeyResponse(key))
}

func apiKeyID(c fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}

func userAPIKey(c fiber.Ctx, userID int, id int) (*ent.APIKey, error) {
	return database.Client.APIKey.Query().
		Where(
			entapikey.ID(id),
			entapikey.UserID(userID),
		).
		Only(c)
}

func apiKeyLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "API key not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load API key")
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Failed to load API key",
	})
}

func apiKeyResponse(key *ent.APIKey) APIKeyResponse {
	return APIKeyResponse{
		ID:         key.ID,
		UserID:     key.UserID,
		Name:       key.Name,
		KeyPreview: keyPreview(key.Key),
		IsActive:   key.IsActive,
		ExpiresAt:  key.ExpiresAt,
		LastUsedAt: key.LastUsedAt,
		CreatedAt:  key.CreatedAt,
		UpdatedAt:  key.UpdatedAt,
	}
}

func keyPreview(key string) string {
	if len(key) <= 12 {
		return "..."
	}
	return key[:8] + "..."
}
