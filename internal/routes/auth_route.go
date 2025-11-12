package routes

import (
	"statio/internal/handlers"
	"statio/internal/middlewares"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes registers all dimension-related routes
func RegisterAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, JWTMiddleware *middlewares.JWTMiddleware) {
	auth := router.Group("/auth")
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.RefreshToken)
	auth.Post("/logout", JWTMiddleware.Protected(), handler.Logout)
	auth.Get("/sso", handler.RedirectSSO)
	auth.Post("/sso", handler.LoginSSO)
}
