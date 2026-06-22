package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"statio/config"
	"statio/internal/dto"
	"statio/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	appConfig  *config.AppConfig
	authConfig *config.AuthConfig
	service    *services.AuthService
	validate   *validator.Validate
}

func NewAuthHandler(appConfig *config.AppConfig, authConfig *config.AuthConfig, service *services.AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		appConfig:  appConfig,
		authConfig: authConfig,
		service:    service,
		validate:   validate,
	}
}

// setRefreshTokenCookie adalah helper function untuk set cookie refresh token
func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, value string, maxAge int) {
	isProd := h.appConfig.AppEnv == "production"

	cookie := &fiber.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "None",
		MaxAge:   maxAge,
	}

	if isProd {
		cookie.Domain = ".databontang.com"
	}

	c.Cookie(cookie)
}

// setStateCookie adalah helper function untuk set cookie state (untuk SSO)
func (h *AuthHandler) setStateCookie(c *fiber.Ctx, value string, maxAge int) {
	isProd := h.appConfig.AppEnv == "production"

	cookie := &fiber.Cookie{
		Name:     "state",
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "None",
		MaxAge:   maxAge,
	}

	if isProd {
		cookie.Domain = ".databontang.com"
	}

	c.Cookie(cookie)
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
			"message": err.Error(),
		})
	}

	// Set refresh token cookie (7 hari)
	h.setRefreshTokenCookie(c, tokens.RefreshToken, 7*24*60*60)

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

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Expire refresh token cookie
	h.setRefreshTokenCookie(c, "", -1)

	return c.JSON(fiber.Map{
		"data":    nil,
		"message": "Logout successful",
	})
}

func generateState() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func (h *AuthHandler) RedirectSSO(c *fiber.Ctx) error {
	state := generateState()

	// Set state cookie (7 hari)
	h.setStateCookie(c, state, 7*24*60*60)

	redirectURL := fmt.Sprintf(
		"%s/api/v1/auth/sso?state=%s&service_id=%s",
		h.authConfig.AuthGateURL,
		state,
		h.authConfig.AuthGateID,
	)

	return c.Redirect(redirectURL)
}

func (h *AuthHandler) LoginSSO(c *fiber.Ctx) error {
	cookieState := c.Cookies("state")
	var payload dto.LoginSSORequest
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

	if payload.State == "" || payload.Token == "" || payload.State != cookieState {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid state or token",
		})
	}

	tokens, err := h.service.LoginBPS(payload.Token)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	// Set refresh token cookie (7 hari)
	h.setRefreshTokenCookie(c, tokens.RefreshToken, 7*24*60*60)

	return c.JSON(fiber.Map{
		"data":    fiber.Map{"access_token": tokens.AccessToken},
		"message": "Login successful",
	})
}

func (h *AuthHandler) LoginInviteToken(c *fiber.Ctx) error {
	var payload dto.LoginInviteTokenRequest
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

	if payload.InviteToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"data":    nil,
			"message": "Invalid invite token",
		})
	}

	tokens, err := h.service.LoginInviteToken(payload.InviteToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"data":    nil,
			"message": err.Error(),
		})
	}

	// Set refresh token cookie (7 hari)
	h.setRefreshTokenCookie(c, tokens.RefreshToken, 7*24*60*60)

	return c.JSON(fiber.Map{
		"data":    fiber.Map{"access_token": tokens.AccessToken},
		"message": "Login successful",
	})
}
