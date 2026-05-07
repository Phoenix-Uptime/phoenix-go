package status_pages

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entstatuspagemonitor "github.com/Phoenix-Uptime/phoenix-go/ent/statuspagemonitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"strconv"
)

type StatusPageMonitorItem struct {
	MonitorID   int     `json:"monitor_id" validate:"required,min=1"`
	DisplayName *string `json:"display_name,omitempty"`
	Weight      *int    `json:"weight,omitempty" validate:"omitempty,min=0"`
	SendURL     *bool   `json:"send_url,omitempty"`
}

type ReplaceStatusPageMonitorsRequest struct {
	Monitors []StatusPageMonitorItem `json:"monitors" validate:"required,dive"`
}

type StatusPageMonitorsResponse struct {
	Monitors []StatusPageMonitorResponse `json:"monitors"`
}

// @Summary List Status Page Monitors
// @Description Returns monitor assignments for one status page owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Status Page ID"
// @Success 200 {object} StatusPageMonitorsResponse "status page monitor list"
// @Failure 400 {object} routes.ErrorResponse "invalid status page id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/{id}/monitors [get]
func ListStatusPageMonitors(c fiber.Ctx) error {
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

	monitors, err := statusPageMonitors(c, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list status page monitors")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list status page monitors",
		})
	}

	return c.Status(fiber.StatusOK).JSON(StatusPageMonitorsResponse{
		Monitors: statusPageMonitorResponses(monitors),
	})
}

// @Summary Replace Status Page Monitors
// @Description Replaces monitor assignments for one status page owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Status Page ID"
// @Param data body ReplaceStatusPageMonitorsRequest true "status page monitor payload"
// @Success 200 {object} StatusPageMonitorsResponse "status page monitor list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/{id}/monitors [put]
func ReplaceStatusPageMonitors(c fiber.Ctx) error {
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

	var req ReplaceStatusPageMonitorsRequest
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

	items, ok := normalizeMonitorItems(req.Monitors)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor ids",
		})
	}
	monitorIDs := make([]int, 0, len(items))
	for _, item := range items {
		monitorIDs = append(monitorIDs, item.MonitorID)
	}
	ok, err = userOwnsMonitorIDs(c, user.ID, monitorIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate status page monitors")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to validate status page monitors",
		})
	}
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor ids",
		})
	}

	tx, err := database.Client.Tx(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start status page monitor transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to replace status page monitors",
		})
	}

	if _, err := tx.StatusPageMonitor.Delete().
		Where(entstatuspagemonitor.StatusPageID(id)).
		Exec(c); err != nil {
		_ = tx.Rollback()
		log.Error().Err(err).Msg("Failed to clear status page monitors")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to replace status page monitors",
		})
	}

	for _, item := range items {
		create := tx.StatusPageMonitor.Create().
			SetStatusPageID(id).
			SetMonitorID(item.MonitorID)
		if item.DisplayName != nil && *item.DisplayName != "" {
			create.SetDisplayName(*item.DisplayName)
		}
		if item.Weight != nil {
			create.SetWeight(*item.Weight)
		}
		if item.SendURL != nil {
			create.SetSendURL(*item.SendURL)
		}
		if _, err := create.Save(c); err != nil {
			_ = tx.Rollback()
			log.Error().Err(err).Msg("Failed to create status page monitor")
			return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Failed to replace status page monitors",
			})
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit status page monitor transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to replace status page monitors",
		})
	}

	monitors, err := statusPageMonitors(c, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload status page monitors")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to replace status page monitors",
		})
	}

	return c.Status(fiber.StatusOK).JSON(StatusPageMonitorsResponse{
		Monitors: statusPageMonitorResponses(monitors),
	})
}

func statusPageMonitors(c fiber.Ctx, statusPageID int) ([]*ent.StatusPageMonitor, error) {
	return database.Client.StatusPageMonitor.Query().
		Where(entstatuspagemonitor.StatusPageID(statusPageID)).
		Order(entstatuspagemonitor.ByWeight()).
		All(c)
}
