package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterOrganizationRoutes registers all organization-related routes
func RegisterOrganizationRoutes(router fiber.Router, handler *handlers.OrganizationHandler) {
	organization := router.Group("/organizations")
	organization.Get("/", handler.GetAllOrganizations)
	organization.Post("/", handler.CreateOrganization)
	organization.Put("/:id", handler.UpdateOrganization)
	organization.Post("/:id/tables", handler.AssignTables)
}
