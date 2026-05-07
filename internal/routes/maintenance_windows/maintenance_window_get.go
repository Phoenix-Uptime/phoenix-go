package maintenance_windows

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmaintenancewindow "github.com/Phoenix-Uptime/phoenix-go/ent/maintenancewindow"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MaintenanceWindowResponse struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	IsActive        bool       `json:"is_active"`
	Strategy        string     `json:"strategy"`
	StartAt         *time.Time `json:"start_at,omitempty"`
	EndAt           *time.Time `json:"end_at,omitempty"`
	Cron            *string    `json:"cron,omitempty"`
	Timezone        *string    `json:"timezone,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	MonitorIDs      []int      `json:"monitor_ids"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// @Summary Get Maintenance Window
// @Description Returns one maintenance window owned by the authenticated user.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Maintenance Window ID"
// @Success 200 {object} MaintenanceWindowResponse "maintenance window"
// @Failure 400 {object} routes.ErrorResponse "invalid maintenance window id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "maintenance window not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows/{id} [get]
func GetMaintenanceWindow(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := maintenanceWindowID(c)
	if err != nil {
		return badRequest(c, "Invalid maintenance window id")
	}

	window, err := userMaintenanceWindow(c, user.ID, id)
	if err != nil {
		return maintenanceWindowLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(maintenanceWindowResponse(window))
}

func maintenanceWindowID(c fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}

func userMaintenanceWindow(c fiber.Ctx, userID int, id int) (*ent.MaintenanceWindow, error) {
	return database.Client.MaintenanceWindow.Query().
		Where(
			entmaintenancewindow.ID(id),
			entmaintenancewindow.UserID(userID),
		).
		WithMonitors().
		Only(c)
}

func maintenanceWindowLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Maintenance window not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load maintenance window")
	return serverError(c, "Failed to load maintenance window")
}

func badRequest(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

func serverError(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

func maintenanceWindowResponse(window *ent.MaintenanceWindow) MaintenanceWindowResponse {
	return MaintenanceWindowResponse{
		ID:              window.ID,
		UserID:          window.UserID,
		Title:           window.Title,
		Description:     window.Description,
		IsActive:        window.IsActive,
		Strategy:        string(window.Strategy),
		StartAt:         window.StartAt,
		EndAt:           window.EndAt,
		Cron:            window.Cron,
		Timezone:        window.Timezone,
		DurationSeconds: window.DurationSeconds,
		MonitorIDs:      monitorIDs(window.Edges.Monitors),
		CreatedAt:       window.CreatedAt,
		UpdatedAt:       window.UpdatedAt,
	}
}

func monitorIDs(monitors []*ent.Monitor) []int {
	ids := make([]int, 0, len(monitors))
	for _, monitor := range monitors {
		ids = append(ids, monitor.ID)
	}
	return ids
}

func normalizeIDs(ids []int) ([]int, bool) {
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if id < 1 {
			return nil, false
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized, true
}

func userOwnsMonitorIDs(c fiber.Ctx, userID int, ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	count, err := database.Client.Monitor.Query().
		Where(
			entmonitor.UserID(userID),
			entmonitor.IDIn(ids...),
		).
		Count(c)
	return count == len(ids), err
}

func validateSchedule(strategy entmaintenancewindow.Strategy, startAt *time.Time, endAt *time.Time, cron *string, durationSeconds *int) error {
	if err := entmaintenancewindow.StrategyValidator(strategy); err != nil {
		return err
	}
	if startAt != nil && endAt != nil && !endAt.After(*startAt) {
		return fmt.Errorf("end_at must be after start_at")
	}
	switch strategy {
	case entmaintenancewindow.StrategySingle:
		if startAt == nil || endAt == nil {
			return fmt.Errorf("start_at and end_at are required for single maintenance windows")
		}
	case entmaintenancewindow.StrategyCron:
		if cron == nil || *cron == "" {
			return fmt.Errorf("cron is required for cron maintenance windows")
		}
		if durationSeconds == nil || *durationSeconds < 1 {
			return fmt.Errorf("duration_seconds is required for cron maintenance windows")
		}
	case entmaintenancewindow.StrategyRecurring:
		if startAt == nil || durationSeconds == nil || *durationSeconds < 1 {
			return fmt.Errorf("start_at and duration_seconds are required for recurring maintenance windows")
		}
	}
	return nil
}
