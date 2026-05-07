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

type CreateMonitorRequest struct {
	Name                string                 `json:"name" validate:"required,min=1"`
	Type                string                 `json:"type" validate:"required,oneof=http keyword json ping tcp smtp dns push grpc"`
	URL                 string                 `json:"url" validate:"required,min=1"`
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

// @Summary Create Monitor
// @Description Creates a monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateMonitorRequest true "monitor payload"
// @Success 201 {object} MonitorResponse "created monitor"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors [post]
func CreateMonitor(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateMonitorRequest
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

	create := database.Client.Monitor.Create().
		SetUserID(user.ID).
		SetName(req.Name).
		SetType(entmonitor.Type(req.Type)).
		SetURL(req.URL)
	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.Interval != nil {
		create.SetInterval(*req.Interval)
	}
	if req.Timeout != nil {
		create.SetTimeout(*req.Timeout)
	}
	if req.Method != nil {
		create.SetMethod(*req.Method)
	}
	if req.AcceptedStatusCodes != nil {
		create.SetAcceptedStatusCodes(req.AcceptedStatusCodes)
	}
	if req.Headers != nil {
		create.SetHeaders(req.Headers)
	}
	if req.Body != nil {
		create.SetBody(*req.Body)
	}
	if req.AuthUsername != nil {
		create.SetAuthUsername(*req.AuthUsername)
	}
	if req.AuthPassword != nil {
		create.SetAuthPassword(*req.AuthPassword)
	}
	if req.FiltersContains != nil {
		create.SetFiltersContains(*req.FiltersContains)
	}
	if req.FiltersNotContains != nil {
		create.SetFiltersNotContains(*req.FiltersNotContains)
	}
	if req.JSONPath != nil {
		create.SetJSONPath(*req.JSONPath)
	}
	if req.ExpectedValue != nil {
		create.SetExpectedValue(*req.ExpectedValue)
	}
	if req.IgnoreTLSErrors != nil {
		create.SetIgnoreTLSErrors(*req.IgnoreTLSErrors)
	}
	if req.MaxRedirects != nil {
		create.SetMaxRedirects(*req.MaxRedirects)
	}
	if req.Retry != nil {
		create.SetRetry(*req.Retry)
	}
	if req.RetryAfter != nil {
		create.SetRetryAfter(*req.RetryAfter)
	}
	if req.PushToken != nil {
		create.SetPushToken(*req.PushToken)
	}
	if req.Config != nil {
		create.SetConfig(req.Config)
	}

	created, err := create.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create monitor")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create monitor",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(monitorResponse(created))
}
