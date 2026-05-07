package tags

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Delete Tag
// @Description Deletes one tag owned by the authenticated user.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Tag ID"
// @Success 200 {object} routes.SuccessResponse "tag deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid tag id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "tag not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /tags/{id} [delete]
func DeleteTag(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := tagID(c)
	if err != nil {
		return badRequest(c, "Invalid tag id")
	}
	if _, err := userTag(c, user.ID, id); err != nil {
		return tagLookupError(c, err)
	}

	if err := database.Client.Tag.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete tag")
		return serverError(c, "Failed to delete tag")
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Tag deleted",
	})
}
