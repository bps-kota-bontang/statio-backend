package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterIntegrationRoutes registers all integration-related routes
func RegisterIntegrationRoutes(router fiber.Router, handler *handlers.IntegrationHandler) {
	integration := router.Group("/integrations")
	integration.Post("/export", handler.ExportDataIntegration)
	integration.Post("/import", handler.ImportDataIntegration)
}
