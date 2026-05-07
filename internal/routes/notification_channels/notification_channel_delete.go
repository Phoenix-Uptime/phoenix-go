package notification_channels

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"strconv"
)

// @Summary Delete Notification Channel
// @Description Deletes one notification channel owned by the authenticated user.
// @Tags Notification Channels
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Notification Channel ID"
// @Success 200 {object} routes.SuccessResponse "notification channel deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid notification channel id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "notification channel not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /notification-channels/{id} [delete]
func DeleteNotificationChannel(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid notification channel id",
		})
	}
	if _, err := userNotificationChannel(c, user.ID, id); err != nil {
		return notificationChannelLookupError(c, err)
	}

	if err := database.Client.NotificationChannel.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete notification channel")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to delete notification channel",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Notification channel deleted",
	})
}
