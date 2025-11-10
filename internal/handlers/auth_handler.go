package handlers

import (
	"statio/config"
	"statio/internal/dto"
	"statio/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	appConfig *config.AppConfig
	service   *services.AuthService
	validate  *validator.Validate
}

func NewAuthHandler(appConfig *config.AppConfig, service *services.AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		appConfig: appConfig,
		service:   service,
		validate:  validate,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var payload dto.LoginRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid request body",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	tokens, err := h.service.Login(&payload)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid credentials",
		})
	}

	isProd := h.appConfig.AppEnv == "production"

	cookie := &fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "None",
		MaxAge:   7 * 24 * 60 * 60,
	}

	if isProd {
		cookie.Domain = ".bpsbontang.com"
	}

	// ✅ Set refresh token ke cookie HttpOnly
	c.Cookie(cookie)

	// kirim access token saja ke frontend
	return c.JSON(fiber.Map{
		"data":    fiber.Map{"access_token": tokens.AccessToken},
		"message": "Login successful",
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{
			"data":    nil,
			"message": "Missing refresh token",
		})
	}

	newAccess, err := h.service.Refresh(refreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    fiber.Map{"access_token": newAccess},
		"message": "Access token refreshed successfully",
	})
}
