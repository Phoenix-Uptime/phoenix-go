package maintenance_windows

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmaintenancewindow "github.com/Phoenix-Uptime/phoenix-go/ent/maintenancewindow"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateMaintenanceWindowRequest struct {
	Title           *string    `json:"title,omitempty" validate:"omitempty,min=1"`
	Description     *string    `json:"description,omitempty"`
	IsActive        *bool      `json:"is_active,omitempty"`
	Strategy        *string    `json:"strategy,omitempty" validate:"omitempty,oneof=manual single recurring cron"`
	StartAt         *time.Time `json:"start_at,omitempty"`
	EndAt           *time.Time `json:"end_at,omitempty"`
	Cron            *string    `json:"cron,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=0"`
	MonitorIDs      *[]int     `json:"monitor_ids,omitempty"`
}

// @Summary Update Maintenance Window
// @Description Updates one maintenance window owned by the authenticated user.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Maintenance Window ID"
// @Param data body UpdateMaintenanceWindowRequest true "maintenance window payload"
// @Success 200 {object} MaintenanceWindowResponse "updated maintenance window"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "maintenance window not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows/{id} [patch]
func UpdateMaintenanceWindow(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := maintenanceWindowID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid maintenance window id",
		})
	}

	current, err := userMaintenanceWindow(c, user.ID, id)
	if err != nil {
		return maintenanceWindowLookupError(c, err)
	}

	var req UpdateMaintenanceWindowRequest
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

	strategy := current.Strategy
	if req.Strategy != nil {
		strategy = entmaintenancewindow.Strategy(*req.Strategy)
	}
	startAt := current.StartAt
	if req.StartAt != nil {
		startAt = req.StartAt
	}
	endAt := current.EndAt
	if req.EndAt != nil {
		endAt = req.EndAt
	}
	cron := current.Cron
	if req.Cron != nil {
		if *req.Cron == "" {
			cron = nil
		} else {
			cron = req.Cron
		}
	}
	durationSeconds := current.DurationSeconds
	if req.DurationSeconds != nil {
		if *req.DurationSeconds == 0 {
			durationSeconds = nil
		} else {
			durationSeconds = req.DurationSeconds
		}
	}
	if err := validateSchedule(strategy, startAt, endAt, cron, durationSeconds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	monitorIDs := monitorIDs(current.Edges.Monitors)
	if req.MonitorIDs != nil {
		monitorIDs = *req.MonitorIDs
	}
	monitorIDs, ok := normalizeIDs(monitorIDs)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor ids",
		})
	}
	ok, err = userOwnsMonitorIDs(c, user.ID, monitorIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate maintenance window monitors")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to validate maintenance window",
		})
	}
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor ids",
		})
	}

	update := database.Client.MaintenanceWindow.UpdateOneID(id).ClearMonitors()
	if req.Title != nil {
		update.SetTitle(*req.Title)
	}
	if req.Description != nil {
		if *req.Description == "" {
			update.ClearDescription()
		} else {
			update.SetDescription(*req.Description)
		}
	}
	if req.IsActive != nil {
		update.SetIsActive(*req.IsActive)
	}
	if req.Strategy != nil {
		update.SetStrategy(strategy)
	}
	if req.StartAt != nil {
		update.SetStartAt(*req.StartAt)
	}
	if req.EndAt != nil {
		update.SetEndAt(*req.EndAt)
	}
	if req.Cron != nil {
		if *req.Cron == "" {
			update.ClearCron()
		} else {
			update.SetCron(*req.Cron)
		}
	}
	if req.Timezone != nil {
		if *req.Timezone == "" {
			update.ClearTimezone()
		} else {
			update.SetTimezone(*req.Timezone)
		}
	}
	if req.DurationSeconds != nil {
		if *req.DurationSeconds == 0 {
			update.ClearDurationSeconds()
		} else {
			update.SetDurationSeconds(*req.DurationSeconds)
		}
	}
	if len(monitorIDs) > 0 {
		update.AddMonitorIDs(monitorIDs...)
	}

	if _, err := update.Save(c); err != nil {
		log.Error().Err(err).Msg("Failed to update maintenance window")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update maintenance window",
		})
	}

	window, err := userMaintenanceWindow(c, user.ID, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload maintenance window")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update maintenance window",
		})
	}

	return c.Status(fiber.StatusOK).JSON(maintenanceWindowResponse(window))
}
