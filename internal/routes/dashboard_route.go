package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterDashboardRoutes registers all dashboard-related routes
func RegisterDashboardRoutes(router fiber.Router, handler *handlers.DashboardHandler) {
	dashboard := router.Group("/dashboard")
	dashboard.Get("/statistics", handler.GetDashboardStatistics)
	dashboard.Get("/organization-completion", handler.GetOrganizationCompletion)
	dashboard.Get("/top-performers", handler.GetTopPerformers)
	dashboard.Get("/organizations-need-attention", handler.GetOrganizationsNeedAttention)
}
