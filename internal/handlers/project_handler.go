package handlers

import (
	"errors"
	"statio/internal/dto"
	"statio/internal/repositories"
	"statio/internal/services"
	"statio/utils"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	service  *services.ProjectService
	validate *validator.Validate
}

func NewProjectHandler(service *services.ProjectService, validate *validator.Validate) *ProjectHandler {
	return &ProjectHandler{service: service, validate: validate}
}

func (h *ProjectHandler) isAdmin(c *fiber.Ctx) bool {
	roles, ok := c.Locals("roles").([]string)
	return ok && utils.IsAdmin(roles)
}

func (h *ProjectHandler) forbidden(c *fiber.Ctx) error {
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"data":    nil,
		"message": "You are not authorized to manage projects",
	})
}

func projectErrorStatus(err error) int {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, repositories.ErrProjectTableExists):
		return fiber.StatusConflict
	default:
		return fiber.StatusInternalServerError
	}
}

func (h *ProjectHandler) projectError(c *fiber.Ctx, err error) error {
	return c.Status(projectErrorStatus(err)).JSON(fiber.Map{
		"data":    nil,
		"message": err.Error(),
	})
}

func (h *ProjectHandler) GetAllProjects(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("per_page", 10)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	projects, total, err := h.service.GetAllPaginated(
		c.Query("search"),
		page,
		perPage,
		c.Query("sort_by", "no"),
		c.Query("sort_order", "asc"),
	)
	if err != nil {
		return h.projectError(c, err)
	}
	return c.JSON(fiber.Map{
		"data":    projects,
		"message": "Projects fetched successfully",
		"meta":    utils.NewPaginationMeta(total, page, perPage),
	})
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	project, err := h.service.GetByID(c.Params("id"))
	if err != nil {
		return h.projectError(c, err)
	}
	return c.JSON(fiber.Map{"data": project, "message": "Project fetched successfully"})
}

func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	var payload dto.CreateProjectRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"data": nil, "message": "Invalid request payload"})
	}
	payload.Name = strings.TrimSpace(payload.Name)
	if err := h.validate.Struct(&payload); err != nil || payload.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"data": nil, "message": "Project name is required"})
	}
	project, err := h.service.Create(&payload)
	if err != nil {
		return h.projectError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": project, "message": "Project created successfully"})
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	var payload dto.UpdateProjectRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"data": nil, "message": "Invalid request payload"})
	}
	payload.Name = strings.TrimSpace(payload.Name)
	if err := h.validate.Struct(&payload); err != nil || payload.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"data": nil, "message": "Project name is required"})
	}
	project, err := h.service.Update(c.Params("id"), &payload)
	if err != nil {
		return h.projectError(c, err)
	}
	return c.JSON(fiber.Map{"data": project, "message": "Project updated successfully"})
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	if err := h.service.Delete(c.Params("id")); err != nil {
		return h.projectError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProjectHandler) AddTable(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	var payload dto.ProjectTableRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"data": nil, "message": "Invalid request payload"})
	}
	if err := h.validateProjectTable(&payload); err != nil {
		return err
	}
	if err := h.service.AddTable(c.Params("id"), &payload); err != nil {
		return h.projectError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": nil, "message": "Table added to project successfully"})
}

func (h *ProjectHandler) UpdateTable(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	var payload dto.ProjectTableRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"data": nil, "message": "Invalid request payload"})
	}
	payload.TableID = c.Params("tableId")
	if err := h.validateProjectTable(&payload); err != nil {
		return err
	}
	if err := h.service.UpdateTable(c.Params("id"), c.Params("tableId"), &payload); err != nil {
		return h.projectError(c, err)
	}
	return c.JSON(fiber.Map{"data": nil, "message": "Project table updated successfully"})
}

func (h *ProjectHandler) RemoveTable(c *fiber.Ctx) error {
	if !h.isAdmin(c) {
		return h.forbidden(c)
	}
	if err := h.service.RemoveTable(c.Params("id"), c.Params("tableId")); err != nil {
		return h.projectError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProjectHandler) validateProjectTable(payload *dto.ProjectTableRequest) error {
	payload.TableID = strings.TrimSpace(payload.TableID)
	payload.Page = strings.TrimSpace(payload.Page)
	payload.TableNumber = strings.TrimSpace(payload.TableNumber)
	if err := h.validate.Struct(payload); err != nil || payload.TableID == "" || payload.Page == "" || payload.TableNumber == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Table, page, and table number are required")
	}
	return nil
}
