package tags

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	enttag "github.com/Phoenix-Uptime/phoenix-go/ent/tag"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type TagResponse struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Color       *string   `json:"color,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// @Summary Get Tag
// @Description Returns one tag owned by the authenticated user.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Tag ID"
// @Success 200 {object} TagResponse "tag"
// @Failure 400 {object} routes.ErrorResponse "invalid tag id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "tag not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /tags/{id} [get]
func GetTag(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := tagID(c)
	if err != nil {
		return badRequest(c, "Invalid tag id")
	}

	tag, err := userTag(c, user.ID, id)
	if err != nil {
		return tagLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(tagResponse(tag))
}

func tagID(c fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}

func userTag(c fiber.Ctx, userID int, id int) (*ent.Tag, error) {
	return database.Client.Tag.Query().
		Where(
			enttag.ID(id),
			enttag.UserID(userID),
		).
		Only(c)
}

func tagLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Tag not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load tag")
	return serverError(c, "Failed to load tag")
}

func badRequest(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: message,
	})
}

func conflict(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(routes.ErrorResponse{
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

func tagResponse(tag *ent.Tag) TagResponse {
	return TagResponse{
		ID:          tag.ID,
		UserID:      tag.UserID,
		Name:        tag.Name,
		Description: tag.Description,
		Color:       tag.Color,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}
