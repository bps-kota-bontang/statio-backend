package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router fiber.Router, handler *handlers.UserHandler) {
	user := router.Group("/users")
	user.Get("/me", handler.Me)
}
