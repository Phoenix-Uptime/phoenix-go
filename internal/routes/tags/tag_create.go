package tags

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreateTagRequest struct {
	Name        string  `json:"name" validate:"required,min=1"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor|rgb|rgba|hsl|hsla"`
}

// @Summary Create Tag
// @Description Creates a tag owned by the authenticated user.
// @Tags Tags
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateTagRequest true "tag payload"
// @Success 201 {object} TagResponse "created tag"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 409 {object} routes.ErrorResponse "tag already exists"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /tags [post]
func CreateTag(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateTagRequest
	if err := c.Bind().Body(&req); err != nil {
		return badRequest(c, "Invalid request payload")
	}
	if err := validator.New().Struct(&req); err != nil {
		return badRequest(c, "Invalid input: "+err.Error())
	}

	create := database.Client.Tag.Create().
		SetUserID(user.ID).
		SetName(req.Name)
	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.Color != nil {
		create.SetColor(*req.Color)
	}

	tag, err := create.Save(c)
	if err != nil {
		if ent.IsConstraintError(err) {
			return conflict(c, "Tag already exists")
		}
		log.Error().Err(err).Msg("Failed to create tag")
		return serverError(c, "Failed to create tag")
	}

	return c.Status(fiber.StatusCreated).JSON(tagResponse(tag))
}
