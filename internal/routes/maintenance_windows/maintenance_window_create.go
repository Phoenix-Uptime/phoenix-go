package maintenance_windows

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmaintenancewindow "github.com/Phoenix-Uptime/phoenix-go/ent/maintenancewindow"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreateMaintenanceWindowRequest struct {
	Title           string     `json:"title" validate:"required,min=1"`
	Description     *string    `json:"description,omitempty"`
	IsActive        *bool      `json:"is_active,omitempty"`
	Strategy        string     `json:"strategy,omitempty" validate:"omitempty,oneof=manual single recurring cron"`
	StartAt         *time.Time `json:"start_at,omitempty"`
	EndAt           *time.Time `json:"end_at,omitempty"`
	Cron            *string    `json:"cron,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty" validate:"omitempty,min=1"`
	MonitorIDs      []int      `json:"monitor_ids,omitempty"`
}

// @Summary Create Maintenance Window
// @Description Creates a maintenance window owned by the authenticated user.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateMaintenanceWindowRequest true "maintenance window payload"
// @Success 201 {object} MaintenanceWindowResponse "created maintenance window"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows [post]
func CreateMaintenanceWindow(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateMaintenanceWindowRequest
	if err := c.Bind().Body(&req); err != nil {
		return badRequest(c, "Invalid request payload")
	}
	if err := validator.New().Struct(&req); err != nil {
		return badRequest(c, "Invalid input: "+err.Error())
	}

	strategy := entmaintenancewindow.StrategySingle
	if req.Strategy != "" {
		strategy = entmaintenancewindow.Strategy(req.Strategy)
	}
	if err := validateSchedule(strategy, req.StartAt, req.EndAt, req.Cron, req.DurationSeconds); err != nil {
		return badRequest(c, err.Error())
	}

	monitorIDs, ok := normalizeIDs(req.MonitorIDs)
	if !ok {
		return badRequest(c, "Invalid monitor ids")
	}
	ok, err := userOwnsMonitorIDs(c, user.ID, monitorIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate maintenance window monitors")
		return serverError(c, "Failed to validate maintenance window")
	}
	if !ok {
		return badRequest(c, "Invalid monitor ids")
	}

	create := database.Client.MaintenanceWindow.Create().
		SetUserID(user.ID).
		SetTitle(req.Title).
		SetStrategy(strategy)
	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.IsActive != nil {
		create.SetIsActive(*req.IsActive)
	}
	if req.StartAt != nil {
		create.SetStartAt(*req.StartAt)
	}
	if req.EndAt != nil {
		create.SetEndAt(*req.EndAt)
	}
	if req.Cron != nil {
		create.SetCron(*req.Cron)
	}
	if req.Timezone != nil {
		create.SetTimezone(*req.Timezone)
	}
	if req.DurationSeconds != nil {
		create.SetDurationSeconds(*req.DurationSeconds)
	}
	if len(monitorIDs) > 0 {
		create.AddMonitorIDs(monitorIDs...)
	}

	window, err := create.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create maintenance window")
		return serverError(c, "Failed to create maintenance window")
	}

	window, err = userMaintenanceWindow(c, user.ID, window.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload maintenance window")
		return serverError(c, "Failed to create maintenance window")
	}

	return c.Status(fiber.StatusCreated).JSON(maintenanceWindowResponse(window))
}
