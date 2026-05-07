package api_keys

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreateAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required,min=1"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateAPIKeyResponse struct {
	APIKey string         `json:"api_key"`
	Key    APIKeyResponse `json:"key"`
}

// @Summary Create API Key
// @Description Creates an API key owned by the authenticated user. The secret is returned only once in this response.
// @Tags API Keys
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateAPIKeyRequest true "API key payload"
// @Success 201 {object} CreateAPIKeyResponse "created API key"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /api-keys [post]
func CreateAPIKey(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateAPIKeyRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}
	if err := validator.New().Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}
	if req.ExpiresAt != nil && !req.ExpiresAt.After(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "API key expiration must be in the future",
		})
	}

	secret, err := generateAPIKey()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate API key")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create API key",
		})
	}

	create := database.Client.APIKey.Create().
		SetUserID(user.ID).
		SetName(req.Name).
		SetKey(secret)
	if req.ExpiresAt != nil {
		create.SetExpiresAt(*req.ExpiresAt)
	}

	key, err := create.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create API key")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create API key",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateAPIKeyResponse{
		APIKey: secret,
		Key:    apiKeyResponse(key),
	})
}

func generateAPIKey() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "phx_" + base64.RawURLEncoding.EncodeToString(raw), nil
}
