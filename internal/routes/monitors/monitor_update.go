package monitors

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateMonitorRequest struct {
	Name                *string                `json:"name,omitempty" validate:"omitempty,min=1"`
	Type                *string                `json:"type,omitempty" validate:"omitempty,oneof=http keyword json ping tcp smtp dns push grpc"`
	URL                 *string                `json:"url,omitempty" validate:"omitempty,min=1"`
	Description         *string                `json:"description,omitempty"`
	Interval            *int                   `json:"interval,omitempty" validate:"omitempty,min=1"`
	Timeout             *int                   `json:"timeout,omitempty" validate:"omitempty,min=1"`
	Method              *string                `json:"method,omitempty" validate:"omitempty,min=1"`
	AcceptedStatusCodes []string               `json:"accepted_status_codes,omitempty"`
	Headers             map[string]string      `json:"headers,omitempty"`
	Body                *string                `json:"body,omitempty"`
	AuthUsername        *string                `json:"auth_username,omitempty"`
	AuthPassword        *string                `json:"auth_password,omitempty"`
	FiltersContains     *string                `json:"filters_contains,omitempty"`
	FiltersNotContains  *string                `json:"filters_not_contains,omitempty"`
	JSONPath            *string                `json:"json_path,omitempty"`
	ExpectedValue       *string                `json:"expected_value,omitempty"`
	IgnoreTLSErrors     *bool                  `json:"ignore_tls_errors,omitempty"`
	MaxRedirects        *int                   `json:"max_redirects,omitempty" validate:"omitempty,min=0"`
	Retry               *int                   `json:"retry,omitempty" validate:"omitempty,min=0"`
	RetryAfter          *int                   `json:"retry_after,omitempty" validate:"omitempty,min=0"`
	PushToken           *string                `json:"push_token,omitempty"`
	Config              map[string]interface{} `json:"config,omitempty"`
}

// @Summary Update Monitor
// @Description Updates one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Param data body UpdateMonitorRequest true "monitor payload"
// @Success 200 {object} MonitorResponse "updated monitor"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id} [patch]
func UpdateMonitor(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor id",
		})
	}
	if _, err := userMonitor(c, user.ID, id); err != nil {
		return monitorLookupError(c, err)
	}

	var req UpdateMonitorRequest
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

	update := database.Client.Monitor.UpdateOneID(id)
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.Type != nil {
		update.SetType(entmonitor.Type(*req.Type))
	}
	if req.URL != nil {
		update.SetURL(*req.URL)
	}
	if req.Description != nil {
		if *req.Description == "" {
			update.ClearDescription()
		} else {
			update.SetDescription(*req.Description)
		}
	}
	if req.Interval != nil {
		update.SetInterval(*req.Interval)
	}
	if req.Timeout != nil {
		update.SetTimeout(*req.Timeout)
	}
	if req.Method != nil {
		update.SetMethod(*req.Method)
	}
	if req.AcceptedStatusCodes != nil {
		update.SetAcceptedStatusCodes(req.AcceptedStatusCodes)
	}
	if req.Headers != nil {
		update.SetHeaders(req.Headers)
	}
	if req.Body != nil {
		if *req.Body == "" {
			update.ClearBody()
		} else {
			update.SetBody(*req.Body)
		}
	}
	if req.AuthUsername != nil {
		if *req.AuthUsername == "" {
			update.ClearAuthUsername()
		} else {
			update.SetAuthUsername(*req.AuthUsername)
		}
	}
	if req.AuthPassword != nil {
		if *req.AuthPassword == "" {
			update.ClearAuthPassword()
		} else {
			update.SetAuthPassword(*req.AuthPassword)
		}
	}
	if req.FiltersContains != nil {
		if *req.FiltersContains == "" {
			update.ClearFiltersContains()
		} else {
			update.SetFiltersContains(*req.FiltersContains)
		}
	}
	if req.FiltersNotContains != nil {
		if *req.FiltersNotContains == "" {
			update.ClearFiltersNotContains()
		} else {
			update.SetFiltersNotContains(*req.FiltersNotContains)
		}
	}
	if req.JSONPath != nil {
		if *req.JSONPath == "" {
			update.ClearJSONPath()
		} else {
			update.SetJSONPath(*req.JSONPath)
		}
	}
	if req.ExpectedValue != nil {
		if *req.ExpectedValue == "" {
			update.ClearExpectedValue()
		} else {
			update.SetExpectedValue(*req.ExpectedValue)
		}
	}
	if req.IgnoreTLSErrors != nil {
		update.SetIgnoreTLSErrors(*req.IgnoreTLSErrors)
	}
	if req.MaxRedirects != nil {
		update.SetMaxRedirects(*req.MaxRedirects)
	}
	if req.Retry != nil {
		update.SetRetry(*req.Retry)
	}
	if req.RetryAfter != nil {
		update.SetRetryAfter(*req.RetryAfter)
	}
	if req.PushToken != nil {
		if *req.PushToken == "" {
			update.ClearPushToken()
		} else {
			update.SetPushToken(*req.PushToken)
		}
	}
	if req.Config != nil {
		update.SetConfig(req.Config)
	}

	updated, err := update.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update monitor")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update monitor",
		})
	}

	return c.Status(fiber.StatusOK).JSON(monitorResponse(updated))
}
