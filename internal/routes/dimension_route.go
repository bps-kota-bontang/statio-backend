package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterDimensionRoutes registers all dimension-related routes
func RegisterDimensionRoutes(router fiber.Router, handler *handlers.DimensionHandler) {
	dimension := router.Group("/dimensions")
	dimension.Get("/", handler.GetAllDimensions)
	dimension.Post("/", handler.CreateDimension)
	dimension.Get("/:id", handler.GetDimension)
	dimension.Put("/:id", handler.UpdateDimension)
}
