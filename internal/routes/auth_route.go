package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes registers all dimension-related routes
func RegisterAuthRoutes(router fiber.Router, handler *handlers.AuthHandler) {
	auth := router.Group("/auth")
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)
	auth.Post("/logout", handler.Logout)
	auth.Get("/sso", handler.RedirectSSO)
	auth.Post("/sso", handler.LoginSSO)
}
