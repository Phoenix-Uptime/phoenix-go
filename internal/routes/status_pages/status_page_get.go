package status_pages

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	entstatuspage "github.com/Phoenix-Uptime/phoenix-go/ent/statuspage"
	entstatuspagemonitor "github.com/Phoenix-Uptime/phoenix-go/ent/statuspagemonitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type StatusPageResponse struct {
	ID                   int                         `json:"id"`
	UserID               int                         `json:"user_id"`
	Slug                 string                      `json:"slug"`
	Name                 string                      `json:"name"`
	Description          *string                     `json:"description,omitempty"`
	IsPublic             bool                        `json:"is_public"`
	HasPassword          bool                        `json:"has_password"`
	Theme                string                      `json:"theme"`
	CustomCSS            *string                     `json:"custom_css,omitempty"`
	FooterText           *string                     `json:"footer_text,omitempty"`
	ShowTags             bool                        `json:"show_tags"`
	ShowCharts           bool                        `json:"show_charts"`
	ShowUptimePercentage bool                        `json:"show_uptime_percentage"`
	ShowPoweredBy        bool                        `json:"show_powered_by"`
	AutoRefreshInterval  int                         `json:"auto_refresh_interval"`
	Monitors             []StatusPageMonitorResponse `json:"monitors"`
	CreatedAt            time.Time                   `json:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at"`
}

type StatusPageMonitorResponse struct {
	ID          int       `json:"id"`
	MonitorID   int       `json:"monitor_id"`
	DisplayName *string   `json:"display_name,omitempty"`
	Weight      int       `json:"weight"`
	SendURL     bool      `json:"send_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// @Summary Get Status Page
// @Description Returns one status page owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Status Page ID"
// @Success 200 {object} StatusPageResponse "status page"
// @Failure 400 {object} routes.ErrorResponse "invalid status page id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/{id} [get]
func GetStatusPage(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid status page id",
		})
	}

	page, err := userStatusPage(c, user.ID, id)
	if err != nil {
		return statusPageLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(statusPageResponse(page))
}

func userStatusPage(c fiber.Ctx, userID int, id int) (*ent.StatusPage, error) {
	return database.Client.StatusPage.Query().
		Where(
			entstatuspage.ID(id),
			entstatuspage.UserID(userID),
		).
		WithStatusPageMonitors(func(query *ent.StatusPageMonitorQuery) {
			query.Order(entstatuspagemonitor.ByWeight())
		}).
		Only(c)
}

func statusPageLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Status page not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load status page")
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Failed to load status page",
	})
}

func statusPageResponse(page *ent.StatusPage) StatusPageResponse {
	return StatusPageResponse{
		ID:                   page.ID,
		UserID:               page.UserID,
		Slug:                 page.Slug,
		Name:                 page.Name,
		Description:          page.Description,
		IsPublic:             page.IsPublic,
		HasPassword:          page.Password != nil && *page.Password != "",
		Theme:                page.Theme,
		CustomCSS:            page.CustomCSS,
		FooterText:           page.FooterText,
		ShowTags:             page.ShowTags,
		ShowCharts:           page.ShowCharts,
		ShowUptimePercentage: page.ShowUptimePercentage,
		ShowPoweredBy:        page.ShowPoweredBy,
		AutoRefreshInterval:  page.AutoRefreshInterval,
		Monitors:             statusPageMonitorResponses(page.Edges.StatusPageMonitors),
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
	}
}

func statusPageMonitorResponses(monitors []*ent.StatusPageMonitor) []StatusPageMonitorResponse {
	responses := make([]StatusPageMonitorResponse, 0, len(monitors))
	for _, monitor := range monitors {
		responses = append(responses, StatusPageMonitorResponse{
			ID:          monitor.ID,
			MonitorID:   monitor.MonitorID,
			DisplayName: monitor.DisplayName,
			Weight:      monitor.Weight,
			SendURL:     monitor.SendURL,
			CreatedAt:   monitor.CreatedAt,
			UpdatedAt:   monitor.UpdatedAt,
		})
	}
	return responses
}

func normalizeMonitorItems(items []StatusPageMonitorItem) ([]StatusPageMonitorItem, bool) {
	seen := make(map[int]struct{}, len(items))
	normalized := make([]StatusPageMonitorItem, 0, len(items))
	for _, item := range items {
		if item.MonitorID < 1 {
			return nil, false
		}
		if _, ok := seen[item.MonitorID]; ok {
			continue
		}
		seen[item.MonitorID] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized, true
}

func userOwnsMonitorIDs(c fiber.Ctx, userID int, monitorIDs []int) (bool, error) {
	if len(monitorIDs) == 0 {
		return true, nil
	}
	count, err := database.Client.Monitor.Query().
		Where(
			entmonitor.UserID(userID),
			entmonitor.IDIn(monitorIDs...),
		).
		Count(c)
	return count == len(monitorIDs), err
}
