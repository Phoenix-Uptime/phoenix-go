package status_pages

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateStatusPageRequest struct {
	Slug                 *string `json:"slug,omitempty" validate:"omitempty,min=1"`
	Name                 *string `json:"name,omitempty" validate:"omitempty,min=1"`
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

// @Summary Update Status Page
// @Description Updates one status page owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Status Page ID"
// @Param data body UpdateStatusPageRequest true "status page payload"
// @Success 200 {object} StatusPageResponse "updated status page"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 409 {object} routes.ErrorResponse "status page slug already exists"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/{id} [patch]
func UpdateStatusPage(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := statusPageID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid status page id",
		})
	}
	if _, err := userStatusPage(c, user.ID, id); err != nil {
		return statusPageLookupError(c, err)
	}

	var req UpdateStatusPageRequest
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

	update := database.Client.StatusPage.UpdateOneID(id)
	if req.Slug != nil {
		slug, ok := normalizeSlug(*req.Slug)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid status page slug",
			})
		}
		update.SetSlug(slug)
	}
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.Description != nil {
		if *req.Description == "" {
			update.ClearDescription()
		} else {
			update.SetDescription(*req.Description)
		}
	}
	if req.IsPublic != nil {
		update.SetIsPublic(*req.IsPublic)
	}
	if req.Password != nil {
		if *req.Password == "" {
			update.ClearPassword()
		} else {
			update.SetPassword(*req.Password)
		}
	}
	if req.Theme != nil {
		update.SetTheme(*req.Theme)
	}
	if req.CustomCSS != nil {
		if *req.CustomCSS == "" {
			update.ClearCustomCSS()
		} else {
			update.SetCustomCSS(*req.CustomCSS)
		}
	}
	if req.FooterText != nil {
		if *req.FooterText == "" {
			update.ClearFooterText()
		} else {
			update.SetFooterText(*req.FooterText)
		}
	}
	if req.ShowTags != nil {
		update.SetShowTags(*req.ShowTags)
	}
	if req.ShowCharts != nil {
		update.SetShowCharts(*req.ShowCharts)
	}
	if req.ShowUptimePercentage != nil {
		update.SetShowUptimePercentage(*req.ShowUptimePercentage)
	}
	if req.ShowPoweredBy != nil {
		update.SetShowPoweredBy(*req.ShowPoweredBy)
	}
	if req.AutoRefreshInterval != nil {
		update.SetAutoRefreshInterval(*req.AutoRefreshInterval)
	}

	if _, err := update.Save(c); err != nil {
		if ent.IsConstraintError(err) {
			return c.Status(fiber.StatusConflict).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Status page slug already exists",
			})
		}
		log.Error().Err(err).Msg("Failed to update status page")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update status page",
		})
	}

	page, err := userStatusPage(c, user.ID, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload status page")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update status page",
		})
	}

	return c.Status(fiber.StatusOK).JSON(statusPageResponse(page))
}
