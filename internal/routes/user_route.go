package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router fiber.Router, handler *handlers.UserHandler) {
	user := router.Group("/users")
	user.Get("/me", handler.Me)
	user.Get("/", handler.GetAllUsers)
	user.Get("/:id", handler.GetUserByID)
	user.Post("/", handler.CreateUser)
	user.Put("/:id", handler.UpdateUser)
	user.Delete("/:id", handler.DeleteUser)
}
