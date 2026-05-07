package monitors

import (
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	enttag "github.com/Phoenix-Uptime/phoenix-go/ent/tag"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MonitorTagResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type MonitorTagsResponse struct {
	Tags []MonitorTagResponse `json:"tags"`
}

type ReplaceMonitorTagsRequest struct {
	TagIDs []int `json:"tag_ids" validate:"required"`
}

// @Summary List Monitor Tags
// @Description Returns tags assigned to one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Success 200 {object} MonitorTagsResponse "monitor tags"
// @Failure 400 {object} routes.ErrorResponse "invalid monitor id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id}/tags [get]
func ListMonitorTags(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return badRequest(c, "Invalid monitor id")
	}

	monitor, err := userMonitor(c, user.ID, id)
	if err != nil {
		return monitorLookupError(c, err)
	}

	tags, err := monitor.QueryTags().
		Order(enttag.ByName(entsql.OrderAsc())).
		All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list monitor tags")
		return serverError(c, "Failed to list monitor tags")
	}

	return c.Status(fiber.StatusOK).JSON(monitorTagsResponse(tags))
}

// @Summary Replace Monitor Tags
// @Description Replaces the full tag set assigned to one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Param data body ReplaceMonitorTagsRequest true "tag IDs"
// @Success 200 {object} MonitorTagsResponse "monitor tags"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id}/tags [put]
func ReplaceMonitorTags(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return badRequest(c, "Invalid monitor id")
	}
	if _, err := userMonitor(c, user.ID, id); err != nil {
		return monitorLookupError(c, err)
	}

	var req ReplaceMonitorTagsRequest
	if err := c.Bind().Body(&req); err != nil {
		return badRequest(c, "Invalid request payload")
	}
	if err := validator.New().Struct(&req); err != nil {
		return badRequest(c, "Invalid input: "+err.Error())
	}

	seen := make(map[int]struct{}, len(req.TagIDs))
	tagIDs := make([]int, 0, len(req.TagIDs))
	for _, tagID := range req.TagIDs {
		if tagID < 1 {
			return badRequest(c, "Invalid tag ids")
		}
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		tagIDs = append(tagIDs, tagID)
	}

	if len(tagIDs) > 0 {
		count, err := database.Client.Tag.Query().
			Where(
				enttag.UserID(user.ID),
				enttag.IDIn(tagIDs...),
			).
			Count(c)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate monitor tags")
			return serverError(c, "Failed to validate monitor tags")
		}
		if count != len(tagIDs) {
			return badRequest(c, "Invalid tag ids")
		}
	}

	update := database.Client.Monitor.UpdateOneID(id).ClearTags()
	if len(tagIDs) > 0 {
		update.AddTagIDs(tagIDs...)
	}
	if err := update.Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to update monitor tags")
		return serverError(c, "Failed to update monitor tags")
	}
	if len(tagIDs) == 0 {
		return c.Status(fiber.StatusOK).JSON(MonitorTagsResponse{Tags: []MonitorTagResponse{}})
	}

	tags, err := database.Client.Tag.Query().
		Where(
			enttag.UserID(user.ID),
			enttag.IDIn(tagIDs...),
		).
		Order(enttag.ByName(entsql.OrderAsc())).
		All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list monitor tags")
		return serverError(c, "Failed to list monitor tags")
	}

	return c.Status(fiber.StatusOK).JSON(monitorTagsResponse(tags))
}

func monitorTagsResponse(tags []*ent.Tag) MonitorTagsResponse {
	response := MonitorTagsResponse{
		Tags: make([]MonitorTagResponse, 0, len(tags)),
	}
	for _, tag := range tags {
		response.Tags = append(response.Tags, MonitorTagResponse{
			ID:          tag.ID,
			Name:        tag.Name,
			Description: tag.Description,
			Color:       tag.Color,
		})
	}
	return response
}
