package tags

import (
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	enttag "github.com/Phoenix-Uptime/phoenix-go/ent/tag"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type TagListResponse struct {
	Tags []TagResponse `json:"tags"`
}

// @Summary List Tags
// @Description Returns tags owned by the authenticated user.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} TagListResponse "tag list"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /tags [get]
func ListTags(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	tags, err := database.Client.Tag.Query().
		Where(enttag.UserID(user.ID)).
		Order(enttag.ByName(entsql.OrderAsc())).
		All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list tags")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list tags",
		})
	}

	response := TagListResponse{
		Tags: make([]TagResponse, 0, len(tags)),
	}
	for _, tag := range tags {
		response.Tags = append(response.Tags, tagResponse(tag))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
