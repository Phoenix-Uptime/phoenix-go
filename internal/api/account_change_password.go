package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=6" swagger:"example=currentpassword123"`
	NewPassword     string `json:"new_password" validate:"required,min=6" swagger:"example=newpassword123"`
}

type PasswordChangeResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// @Summary Change User Password
// @Description Allows an authenticated user to change their password.
// @Tags Account
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body PasswordChangeRequest true "Password change data"
// @Success 200 {object} PasswordChangeResponse "Password changed successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid input"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid or missing API key"
// @Failure 403 {object} ErrorResponse "Forbidden - incorrect current password"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/change-password [post]
func ChangePassword(c *fiber.Ctx) error {
	var req PasswordChangeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}

	user := c.Locals("user").(*models.User)

	// Check if the current password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Status:  "error",
			Message: "Incorrect current password",
		})
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash new password")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	// Update the user's password
	user.Password = string(hashedPassword)
	if err := models.DB.Save(user).Error; err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(PasswordChangeResponse{
		Status:  "success",
		Message: "Password changed successfully",
	})
}
