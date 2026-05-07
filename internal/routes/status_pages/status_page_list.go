package status_pages

import (
	"strconv"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entstatuspage "github.com/Phoenix-Uptime/phoenix-go/ent/statuspage"
	entstatuspagemonitor "github.com/Phoenix-Uptime/phoenix-go/ent/statuspagemonitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type StatusPageListResponse struct {
	StatusPages []StatusPageResponse `json:"status_pages"`
}

// @Summary List Status Pages
// @Description Returns status pages owned by the authenticated user.
// @Tags Status Pages
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param is_public query bool false "filter by public/private status"
// @Success 200 {object} StatusPageListResponse "status page list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages [get]
func ListStatusPages(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	query := database.Client.StatusPage.Query().
		Where(entstatuspage.UserID(user.ID)).
		WithStatusPageMonitors(func(query *ent.StatusPageMonitorQuery) {
			query.Order(entstatuspagemonitor.ByWeight())
		}).
		Order(entstatuspage.ByCreatedAt(entsql.OrderDesc()))

	if isPublic := c.Query("is_public"); isPublic != "" {
		value, err := strconv.ParseBool(isPublic)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid is_public filter",
			})
		}
		query.Where(entstatuspage.IsPublic(value))
	}

	pages, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list status pages")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list status pages",
		})
	}

	response := StatusPageListResponse{
		StatusPages: make([]StatusPageResponse, 0, len(pages)),
	}
	for _, page := range pages {
		response.StatusPages = append(response.StatusPages, statusPageResponse(page))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
