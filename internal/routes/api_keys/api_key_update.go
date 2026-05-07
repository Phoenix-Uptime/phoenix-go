package api_keys

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateAPIKeyRequest struct {
	Name           *string    `json:"name,omitempty" validate:"omitempty,min=1"`
	IsActive       *bool      `json:"is_active,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	ClearExpiresAt *bool      `json:"clear_expires_at,omitempty"`
}

// @Summary Update API Key
// @Description Updates API key metadata owned by the authenticated user without rotating the secret.
// @Tags API Keys
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "API Key ID"
// @Param data body UpdateAPIKeyRequest true "API key payload"
// @Success 200 {object} APIKeyResponse "updated API key"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "API key not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /api-keys/{id} [patch]
func UpdateAPIKey(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid API key id",
		})
	}
	if _, err := userAPIKey(c, user.ID, id); err != nil {
		return apiKeyLookupError(c, err)
	}

	var req UpdateAPIKeyRequest
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

	update := database.Client.APIKey.UpdateOneID(id)
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.IsActive != nil {
		update.SetIsActive(*req.IsActive)
	}
	if req.ClearExpiresAt != nil && *req.ClearExpiresAt {
		update.ClearExpiresAt()
	} else if req.ExpiresAt != nil {
		update.SetExpiresAt(*req.ExpiresAt)
	}

	key, err := update.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update API key")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update API key",
		})
	}

	return c.Status(fiber.StatusOK).JSON(apiKeyResponse(key))
}
