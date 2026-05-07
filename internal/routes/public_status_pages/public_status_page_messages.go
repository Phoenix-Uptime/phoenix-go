package public_status_pages

import (
	"strconv"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entstatusmessage "github.com/Phoenix-Uptime/phoenix-go/ent/statusmessage"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type PublicStatusPageMessageResponse struct {
	ID         int       `json:"id"`
	IncidentID *int      `json:"incident_id,omitempty"`
	ParentID   *int      `json:"parent_id,omitempty"`
	Type       string    `json:"type"`
	Title      *string   `json:"title,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PublicStatusPageMessagesResponse struct {
	Messages []PublicStatusPageMessageResponse `json:"messages"`
}

// @Summary List Public Status Page Messages
// @Description Returns messages for a public status page by slug without API key authentication.
// @Tags Public Status Pages
// @Accept json
// @Produce json
// @Param slug path string true "Status Page Slug"
// @Param password query string false "Status page password when protected"
// @Param limit query int false "Maximum messages to return"
// @Success 200 {object} PublicStatusPageMessagesResponse "public message list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "password required or invalid"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/public/{slug}/messages [get]
func ListPublicStatusPageMessages(c fiber.Ctx) error {
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

	limit := 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			return badRequest(c, "Invalid limit")
		}
		limit = parsedLimit
	}
	if limit > 200 {
		limit = 200
	}

	messages, err := database.Client.StatusMessage.Query().
		Where(entstatusmessage.StatusPageID(page.ID)).
		Order(entstatusmessage.ByCreatedAt(entsql.OrderDesc())).
		Limit(limit).
		All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list public status page messages")
		return serverError(c, "Failed to list status page messages")
	}

	response := PublicStatusPageMessagesResponse{
		Messages: make([]PublicStatusPageMessageResponse, 0, len(messages)),
	}
	for _, message := range messages {
		response.Messages = append(response.Messages, publicStatusMessageResponse(message))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func publicStatusMessageResponse(message *ent.StatusMessage) PublicStatusPageMessageResponse {
	return PublicStatusPageMessageResponse{
		ID:         message.ID,
		IncidentID: message.IncidentID,
		ParentID:   message.ParentID,
		Type:       string(message.Type),
		Title:      message.Title,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
		UpdatedAt:  message.UpdatedAt,
	}
}
