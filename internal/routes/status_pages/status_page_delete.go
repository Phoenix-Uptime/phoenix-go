package status_pages

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"strconv"
)

// @Summary Delete Status Page
// @Description Deletes one status page owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Status Page ID"
// @Success 200 {object} routes.SuccessResponse "status page deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid status page id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/{id} [delete]
func DeleteStatusPage(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid status page id",
		})
	}
	if _, err := userStatusPage(c, user.ID, id); err != nil {
		return statusPageLookupError(c, err)
	}

	if err := database.Client.StatusPage.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete status page")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to delete status page",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Status page deleted",
	})
}
