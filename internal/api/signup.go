package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	// Username must be between 3 to 20 characters long, allowing only alphanumeric characters,
	// underscores, and hyphens. It is required and should be unique.
	Username string `json:"username" validate:"required,min=3,max=20,alphanumunicode" swagger:"example=example_user"`
	Email    string `json:"email" validate:"required,email" swagger:"example=user@example.com"`
	Password string `json:"password" validate:"required,min=6" swagger:"example=examplepassword"`
}

type SignupResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// @Summary User Signup
// @Description Allows a new user to sign up, but only one user can exist.
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body SignupRequest true "User signup data"
// @Success 201 {object} SignupResponse "User successfully created"
// @Failure 403 {object} ErrorResponse "Signup not allowed if a user already exists"
// @Failure 400 {object} ErrorResponse "Invalid request payload"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /signup [post]
func Signup(c *fiber.Ctx) error {
	var req SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Validation failed: " + validationErrors.Error(),
		})
	}

	var userCount int64
	if err := models.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		log.Error().Err(err).Msg("Failed to check user count")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	// Prevent further signups if a user already exists
	if userCount > 0 {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Status:  "error",
			Message: "User already exists. Signup not allowed.",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	// Generate a new API key using UUID
	apiKey := uuid.New().String()

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		ApiKey:   apiKey,
	}
	if err := models.DB.Create(&user).Error; err != nil {
		log.Error().Err(err).Msg("Failed to create user")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(SignupResponse{
		Status:  "success",
		Message: "User signed up successfully",
	})
}
