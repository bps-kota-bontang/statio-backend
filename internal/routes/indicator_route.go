package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterIndicatorRoutes registers all indicator-related routes
func RegisterIndicatorRoutes(router fiber.Router, handler *handlers.IndicatorHandler) {
	indicator := router.Group("/indicators")
	indicator.Get("/", handler.GetAllIndicators)
	indicator.Post("/", handler.CreateIndicator)
	indicator.Get("/:id", handler.GetIndicator)
	indicator.Put("/:id", handler.UpdateIndicator)
}
