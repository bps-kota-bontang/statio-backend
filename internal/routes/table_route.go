package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterTableRoutes registers all table-related routes
func RegisterTableRoutes(router fiber.Router, handler *handlers.TableHandler) {
	table := router.Group("/tables")
	table.Get("/", handler.GetAllTables)
	table.Post("/", handler.CreateTable)
	table.Get("/labels", handler.GetTableLabels)
	table.Get("/:id", handler.GetTable)
	table.Put("/:id", handler.UpdateTable)
	table.Put("/:id/facts", handler.UpdateFacts)
	table.Put("/:id/labels", handler.UpdateLabels)
	table.Put("/:id/name", handler.UpdateTableName)
	table.Put("/:id/notes", handler.UpdateTableNotes)
	table.Patch("/labels", handler.AddLabelsToTables)
}
