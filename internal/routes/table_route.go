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
	table.Post("/analyze", handler.AnalyzeTables)
	table.Get("/:id", handler.GetTable)
	table.Put("/:id", handler.UpdateTable)
	table.Put("/:id/facts", handler.UpdateFacts)
	table.Post("/:id/analyze", handler.AnalyzeTable)
	table.Get("/:id/facts/missing", handler.GetMissingFacts)
	table.Put("/:id/labels", handler.UpdateLabels)
	table.Put("/:id/name", handler.UpdateTableName)
	table.Put("/:id/notes", handler.UpdateTableNotes)
	table.Put("/:id/lock", handler.UpdateTableIsLocked)
	table.Put("/:id/status", handler.UpdateTableStatus)
	table.Patch("/labels", handler.AddLabelsToTables)
}
