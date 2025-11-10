package middlewares

import (
	"statio/internal/services"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type JWTMiddleware struct {
	jwtService *services.JWTService
}

func NewJWTMiddleware(jwtService *services.JWTService) *JWTMiddleware {
	return &JWTMiddleware{jwtService}
}

func (m *JWTMiddleware) Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		userID, roles, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{
					"data":    nil,
					"message": err.Error(),
				},
			)
		}

		c.Locals("user_id", userID)
		c.Locals("roles", roles)
		return c.Next()
	}
}
