package routes

import (
	"statio/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterProjectRoutes(router fiber.Router, handler *handlers.ProjectHandler) {
	project := router.Group("/projects")
	project.Get("/", handler.GetAllProjects)
	project.Post("/", handler.CreateProject)
	project.Get("/:id", handler.GetProject)
	project.Put("/:id", handler.UpdateProject)
	project.Delete("/:id", handler.DeleteProject)
	project.Post("/:id/tables", handler.AddTable)
	project.Put("/:id/tables/:tableId", handler.UpdateTable)
	project.Delete("/:id/tables/:tableId", handler.RemoveTable)
}
