package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterConfigurationRoutes registers all configuration-related routes
func RegisterConfigurationRoutes(router fiber.Router, handler *handlers.ConfigurationHandler) {
	configuration := router.Group("/configurations")
	configuration.Get("/:key", handler.GetConfiguration)
	configuration.Put("/:key", handler.UpdateConfiguration)
}
