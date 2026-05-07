package tags

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateTagRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor|rgb|rgba|hsl|hsla"`
}

// @Summary Update Tag
// @Description Updates one tag owned by the authenticated user.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Tag ID"
// @Param data body UpdateTagRequest true "tag payload"
// @Success 200 {object} TagResponse "updated tag"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "tag not found"
// @Failure 409 {object} routes.ErrorResponse "tag already exists"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /tags/{id} [patch]
func UpdateTag(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := tagID(c)
	if err != nil {
		return badRequest(c, "Invalid tag id")
	}
	if _, err := userTag(c, user.ID, id); err != nil {
		return tagLookupError(c, err)
	}

	var req UpdateTagRequest
	if err := c.Bind().Body(&req); err != nil {
		return badRequest(c, "Invalid request payload")
	}
	if err := validator.New().Struct(&req); err != nil {
		return badRequest(c, "Invalid input: "+err.Error())
	}

	update := database.Client.Tag.UpdateOneID(id)
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
	if req.Color != nil {
		if *req.Color == "" {
			update.ClearColor()
		} else {
			update.SetColor(*req.Color)
		}
	}

	tag, err := update.Save(c)
	if err != nil {
		if ent.IsConstraintError(err) {
			return conflict(c, "Tag already exists")
		}
		log.Error().Err(err).Msg("Failed to update tag")
		return serverError(c, "Failed to update tag")
	}

	return c.Status(fiber.StatusOK).JSON(tagResponse(tag))
}
