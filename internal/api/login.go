package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20,alphanumunicode" swagger:"example=exampleuser"`
	Password string `json:"password" validate:"required,min=6" swagger:"example=examplepassword"`
}

type LoginResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ApiKey  string `json:"api_key,omitempty"`
}

// @Summary User Login
// @Description Logs in a user with username and password, returning an API key if successful.
// @Tags Auth
// @Accept json
// @Produce json
// @Param data body LoginRequest true "User login data"
// @Success 200 {object} LoginResponse "user successfully logged in"
// @Failure 400 {object} ErrorResponse "invalid login payload"
// @Failure 401 {object} ErrorResponse "invalid credentials"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /login [post]
func Login(c *fiber.Ctx) error {
	var req LoginRequest
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
			Message: "Invalid username or password format",
		})
	}

	var user models.User
	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		log.Error().Err(err).Msg("User not found")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
	}

	return c.Status(fiber.StatusOK).JSON(LoginResponse{
		Status:  "success",
		Message: "Login successful",
		ApiKey:  user.ApiKey,
	})
}
