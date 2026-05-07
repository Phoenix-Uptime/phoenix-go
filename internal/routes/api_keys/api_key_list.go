package api_keys

import (
	"strconv"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entapikey "github.com/Phoenix-Uptime/phoenix-go/ent/apikey"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type APIKeyListResponse struct {
	APIKeys []APIKeyResponse `json:"api_keys"`
}

// @Summary List API Keys
// @Description Returns API keys owned by the authenticated user without exposing secrets.
// @Tags API Keys
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param is_active query bool false "filter by active status"
// @Success 200 {object} APIKeyListResponse "API key list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /api-keys [get]
func ListAPIKeys(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	query := database.Client.APIKey.Query().
		Where(entapikey.UserID(user.ID)).
		Order(entapikey.ByCreatedAt(entsql.OrderDesc()))

	if isActive := c.Query("is_active"); isActive != "" {
		value, err := strconv.ParseBool(isActive)
		if err != nil {
			return badRequest(c, "Invalid is_active filter")
		}
		query.Where(entapikey.IsActive(value))
	}

	keys, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list API keys")
		return serverError(c, "Failed to list API keys")
	}

	response := APIKeyListResponse{
		APIKeys: make([]APIKeyResponse, 0, len(keys)),
	}
	for _, key := range keys {
		response.APIKeys = append(response.APIKeys, apiKeyResponse(key))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
