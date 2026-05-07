package public_status_pages

import (
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	entmonitorcheck "github.com/Phoenix-Uptime/phoenix-go/ent/monitorcheck"
	entmonitorstat "github.com/Phoenix-Uptime/phoenix-go/ent/monitorstat"
	entstatuspage "github.com/Phoenix-Uptime/phoenix-go/ent/statuspage"
	entstatuspagemonitor "github.com/Phoenix-Uptime/phoenix-go/ent/statuspagemonitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type PublicStatusPageResponse struct {
	Slug                 string                            `json:"slug"`
	Name                 string                            `json:"name"`
	Description          *string                           `json:"description,omitempty"`
	Status               string                            `json:"status"`
	Theme                string                            `json:"theme"`
	CustomCSS            *string                           `json:"custom_css,omitempty"`
	FooterText           *string                           `json:"footer_text,omitempty"`
	ShowTags             bool                              `json:"show_tags"`
	ShowCharts           bool                              `json:"show_charts"`
	ShowUptimePercentage bool                              `json:"show_uptime_percentage"`
	ShowPoweredBy        bool                              `json:"show_powered_by"`
	AutoRefreshInterval  int                               `json:"auto_refresh_interval"`
	Monitors             []PublicStatusPageMonitorResponse `json:"monitors"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}

type PublicStatusPageMonitorResponse struct {
	MonitorID           int        `json:"monitor_id"`
	Name                string     `json:"name"`
	Description         *string    `json:"description,omitempty"`
	URL                 *string    `json:"url,omitempty"`
	Status              string     `json:"status"`
	Type                string     `json:"type"`
	Weight              int        `json:"weight"`
	LastCheckedAt       *time.Time `json:"last_checked_at,omitempty"`
	ResponseTimeMs      *int       `json:"response_time_ms,omitempty"`
	UptimePercentage    *float64   `json:"uptime_percentage,omitempty"`
	AvgResponseTimeMs   *int       `json:"avg_response_time_ms,omitempty"`
	MaintenanceChecks   *int       `json:"maintenance_checks,omitempty"`
	DowntimeSeconds     *int       `json:"downtime_seconds,omitempty"`
	MonitorCreatedAt    time.Time  `json:"monitor_created_at"`
	AssignmentCreatedAt time.Time  `json:"assignment_created_at"`
}

// @Summary Get Public Status Page
// @Description Returns a public status page by slug without API key authentication.
// @Tags Public Status Pages
// @Accept json
// @Produce json
// @Param slug path string true "Status Page Slug"
// @Param password query string false "Status page password when protected"
// @Success 200 {object} PublicStatusPageResponse "public status page"
// @Failure 401 {object} routes.ErrorResponse "password required or invalid"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/public/{slug} [get]
func GetPublicStatusPage(c fiber.Ctx) error {
	page, err := publicStatusPage(c, c.Params("slug"))
	if err != nil {
		return publicStatusPageLookupError(c, err)
	}
	authorized, err := authorizePublicStatusPage(c, page)
	if err != nil {
		return err
	}
	if !authorized {
		return publicStatusPagePasswordError(c)
	}

	response, err := publicStatusPageResponse(c, page)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build public status page response")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to load status page",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func publicStatusPage(c fiber.Ctx, slug string) (*ent.StatusPage, error) {
	return database.Client.StatusPage.Query().
		Where(
			entstatuspage.Slug(slug),
			entstatuspage.IsPublic(true),
		).
		Select(entstatuspage.Columns...).
		WithStatusPageMonitors(func(query *ent.StatusPageMonitorQuery) {
			query.
				WithMonitor().
				Order(entstatuspagemonitor.ByWeight())
		}).
		Only(c)
}

func authorizePublicStatusPage(c fiber.Ctx, page *ent.StatusPage) (bool, error) {
	protected, err := database.Client.StatusPage.Query().
		Where(
			entstatuspage.ID(page.ID),
			entstatuspage.PasswordNotNil(),
		).
		Exist(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check status page password")
		return false, c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to load status page",
		})
	}
	if !protected {
		return true, nil
	}
	password := c.Get("x-status-page-password")
	if password == "" {
		password = c.Query("password")
	}
	valid, err := database.Client.StatusPage.Query().
		Where(
			entstatuspage.ID(page.ID),
			entstatuspage.Password(password),
		).
		Exist(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate status page password")
		return false, c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to load status page",
		})
	}
	return valid, nil
}

func publicStatusPagePasswordError(c fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Status page password is required",
	})
}

func publicStatusPageLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Status page not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load public status page")
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Failed to load status page",
	})
}

func publicStatusPageResponse(c fiber.Ctx, page *ent.StatusPage) (PublicStatusPageResponse, error) {
	monitors, err := publicStatusPageMonitorResponses(c, page)
	if err != nil {
		return PublicStatusPageResponse{}, err
	}

	return PublicStatusPageResponse{
		Slug:                 page.Slug,
		Name:                 page.Name,
		Description:          page.Description,
		Status:               publicStatus(monitors),
		Theme:                page.Theme,
		CustomCSS:            page.CustomCSS,
		FooterText:           page.FooterText,
		ShowTags:             page.ShowTags,
		ShowCharts:           page.ShowCharts,
		ShowUptimePercentage: page.ShowUptimePercentage,
		ShowPoweredBy:        page.ShowPoweredBy,
		AutoRefreshInterval:  page.AutoRefreshInterval,
		Monitors:             monitors,
		CreatedAt:            page.CreatedAt,
		UpdatedAt:            page.UpdatedAt,
	}, nil
}

func publicStatusPageMonitorResponses(c fiber.Ctx, page *ent.StatusPage) ([]PublicStatusPageMonitorResponse, error) {
	responses := make([]PublicStatusPageMonitorResponse, 0, len(page.Edges.StatusPageMonitors))
	for _, assignment := range page.Edges.StatusPageMonitors {
		monitor := assignment.Edges.Monitor
		if monitor == nil {
			continue
		}

		response := PublicStatusPageMonitorResponse{
			MonitorID:           monitor.ID,
			Name:                publicMonitorName(assignment, monitor),
			Description:         monitor.Description,
			Status:              string(monitor.Status),
			Type:                string(monitor.Type),
			Weight:              assignment.Weight,
			MonitorCreatedAt:    monitor.CreatedAt,
			AssignmentCreatedAt: assignment.CreatedAt,
		}
		if assignment.SendURL {
			response.URL = &monitor.URL
		}

		check, err := latestMonitorCheck(c, monitor.ID)
		if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
		if check != nil {
			response.LastCheckedAt = &check.CheckedAt
			response.ResponseTimeMs = &check.ResponseTimeMs
		}

		if page.ShowUptimePercentage || page.ShowCharts {
			stat, err := latestMonitorStat(c, monitor.ID)
			if err != nil && !ent.IsNotFound(err) {
				return nil, err
			}
			if stat != nil {
				response.UptimePercentage = &stat.UptimePercentage
				response.AvgResponseTimeMs = &stat.AvgResponseTimeMs
				response.MaintenanceChecks = &stat.MaintenanceChecks
				response.DowntimeSeconds = &stat.DowntimeSeconds
			}
		}

		responses = append(responses, response)
	}
	return responses, nil
}

func publicMonitorName(assignment *ent.StatusPageMonitor, monitor *ent.Monitor) string {
	if assignment.DisplayName != nil && *assignment.DisplayName != "" {
		return *assignment.DisplayName
	}
	return monitor.Name
}

func latestMonitorCheck(c fiber.Ctx, monitorID int) (*ent.MonitorCheck, error) {
	return database.Client.MonitorCheck.Query().
		Where(entmonitorcheck.MonitorID(monitorID)).
		Order(entmonitorcheck.ByCheckedAt(entsql.OrderDesc())).
		First(c)
}

func latestMonitorStat(c fiber.Ctx, monitorID int) (*ent.MonitorStat, error) {
	return database.Client.MonitorStat.Query().
		Where(
			entmonitorstat.MonitorID(monitorID),
			entmonitorstat.PeriodEQ(entmonitorstat.PeriodDay),
		).
		Order(entmonitorstat.ByPeriodStart(entsql.OrderDesc())).
		First(c)
}

func publicStatus(monitors []PublicStatusPageMonitorResponse) string {
	if len(monitors) == 0 {
		return "unknown"
	}
	hasDown := false
	hasMaintenance := false
	hasUnknown := false
	for _, monitor := range monitors {
		switch entmonitor.Status(monitor.Status) {
		case entmonitor.StatusDown:
			hasDown = true
		case entmonitor.StatusMaintenance, entmonitor.StatusPaused:
			hasMaintenance = true
		case entmonitor.StatusPending, entmonitor.StatusUnknown:
			hasUnknown = true
		}
	}
	switch {
	case hasDown:
		return "down"
	case hasMaintenance:
		return "maintenance"
	case hasUnknown:
		return "unknown"
	default:
		return "up"
	}
}
