package status_pages

import (
	"strings"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreateStatusPageRequest struct {
	Slug                 string  `json:"slug" validate:"required,min=1"`
	Name                 string  `json:"name" validate:"required,min=1"`
	Description          *string `json:"description,omitempty"`
	IsPublic             *bool   `json:"is_public,omitempty"`
	Password             *string `json:"password,omitempty"`
	Theme                *string `json:"theme,omitempty" validate:"omitempty,min=1"`
	CustomCSS            *string `json:"custom_css,omitempty"`
	FooterText           *string `json:"footer_text,omitempty"`
	ShowTags             *bool   `json:"show_tags,omitempty"`
	ShowCharts           *bool   `json:"show_charts,omitempty"`
	ShowUptimePercentage *bool   `json:"show_uptime_percentage,omitempty"`
	ShowPoweredBy        *bool   `json:"show_powered_by,omitempty"`
	AutoRefreshInterval  *int    `json:"auto_refresh_interval,omitempty" validate:"omitempty,min=30"`
}

// @Summary Create Status Page
// @Description Creates a status page owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateStatusPageRequest true "status page payload"
// @Success 201 {object} StatusPageResponse "created status page"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 409 {object} routes.ErrorResponse "status page slug already exists"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages [post]
func CreateStatusPage(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateStatusPageRequest
	if err := c.Bind().Body(&req); err != nil {
		return badRequest(c, "Invalid request payload")
	}
	if err := validator.New().Struct(&req); err != nil {
		return badRequest(c, "Invalid input: "+err.Error())
	}

	slug, ok := normalizeSlug(req.Slug)
	if !ok {
		return badRequest(c, "Invalid status page slug")
	}

	create := database.Client.StatusPage.Create().
		SetUserID(user.ID).
		SetSlug(slug).
		SetName(req.Name)
	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.IsPublic != nil {
		create.SetIsPublic(*req.IsPublic)
	}
	if req.Password != nil && *req.Password != "" {
		create.SetPassword(*req.Password)
	}
	if req.Theme != nil {
		create.SetTheme(*req.Theme)
	}
	if req.CustomCSS != nil {
		create.SetCustomCSS(*req.CustomCSS)
	}
	if req.FooterText != nil {
		create.SetFooterText(*req.FooterText)
	}
	if req.ShowTags != nil {
		create.SetShowTags(*req.ShowTags)
	}
	if req.ShowCharts != nil {
		create.SetShowCharts(*req.ShowCharts)
	}
	if req.ShowUptimePercentage != nil {
		create.SetShowUptimePercentage(*req.ShowUptimePercentage)
	}
	if req.ShowPoweredBy != nil {
		create.SetShowPoweredBy(*req.ShowPoweredBy)
	}
	if req.AutoRefreshInterval != nil {
		create.SetAutoRefreshInterval(*req.AutoRefreshInterval)
	}

	page, err := create.Save(c)
	if err != nil {
		if ent.IsConstraintError(err) {
			return conflict(c, "Status page slug already exists")
		}
		log.Error().Err(err).Msg("Failed to create status page")
		return serverError(c, "Failed to create status page")
	}

	page, err = userStatusPage(c, user.ID, page.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload status page")
		return serverError(c, "Failed to create status page")
	}

	return c.Status(fiber.StatusCreated).JSON(statusPageResponse(page))
}

func normalizeSlug(slug string) (string, bool) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug == "" {
		return "", false
	}
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return "", false
	}
	return slug, true
}
