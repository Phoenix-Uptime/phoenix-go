package monitors

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MonitorResponse struct {
	ID                  int                    `json:"id"`
	UserID              int                    `json:"user_id"`
	Name                string                 `json:"name"`
	Description         *string                `json:"description,omitempty"`
	URL                 string                 `json:"url"`
	Interval            int                    `json:"interval"`
	Timeout             int                    `json:"timeout"`
	Status              string                 `json:"status"`
	Type                string                 `json:"type"`
	IsActive            bool                   `json:"is_active"`
	Method              string                 `json:"method"`
	AcceptedStatusCodes []string               `json:"accepted_status_codes,omitempty"`
	Headers             map[string]string      `json:"headers,omitempty"`
	Body                *string                `json:"body,omitempty"`
	AuthUsername        *string                `json:"auth_username,omitempty"`
	FiltersContains     *string                `json:"filters_contains,omitempty"`
	FiltersNotContains  *string                `json:"filters_not_contains,omitempty"`
	JSONPath            *string                `json:"json_path,omitempty"`
	ExpectedValue       *string                `json:"expected_value,omitempty"`
	IgnoreTLSErrors     bool                   `json:"ignore_tls_errors"`
	MaxRedirects        int                    `json:"max_redirects"`
	Retry               int                    `json:"retry"`
	RetryAfter          int                    `json:"retry_after"`
	PushToken           *string                `json:"push_token,omitempty"`
	Config              map[string]interface{} `json:"config,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// @Summary Get Monitor
// @Description Returns one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Success 200 {object} MonitorResponse "monitor"
// @Failure 400 {object} routes.ErrorResponse "invalid monitor id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id} [get]
func GetMonitor(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return badRequest(c, "Invalid monitor id")
	}

	monitor, err := userMonitor(c, user.ID, id)
	if err != nil {
		return monitorLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(monitorResponse(monitor))
}

func monitorID(c fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}

func userMonitor(c fiber.Ctx, userID int, id int) (*ent.Monitor, error) {
	return database.Client.Monitor.Query().
		Where(
			entmonitor.ID(id),
			entmonitor.UserID(userID),
		).
		Only(c)
}

func monitorLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Monitor not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load monitor")
	return serverError(c, "Failed to load monitor")
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

func monitorResponse(monitor *ent.Monitor) MonitorResponse {
	return MonitorResponse{
		ID:                  monitor.ID,
		UserID:              monitor.UserID,
		Name:                monitor.Name,
		Description:         monitor.Description,
		URL:                 monitor.URL,
		Interval:            monitor.Interval,
		Timeout:             monitor.Timeout,
		Status:              string(monitor.Status),
		Type:                string(monitor.Type),
		IsActive:            monitor.IsActive,
		Method:              monitor.Method,
		AcceptedStatusCodes: monitor.AcceptedStatusCodes,
		Headers:             monitor.Headers,
		Body:                monitor.Body,
		AuthUsername:        monitor.AuthUsername,
		FiltersContains:     monitor.FiltersContains,
		FiltersNotContains:  monitor.FiltersNotContains,
		JSONPath:            monitor.JSONPath,
		ExpectedValue:       monitor.ExpectedValue,
		IgnoreTLSErrors:     monitor.IgnoreTLSErrors,
		MaxRedirects:        monitor.MaxRedirects,
		Retry:               monitor.Retry,
		RetryAfter:          monitor.RetryAfter,
		PushToken:           monitor.PushToken,
		Config:              monitor.Config,
		CreatedAt:           monitor.CreatedAt,
		UpdatedAt:           monitor.UpdatedAt,
	}
}
