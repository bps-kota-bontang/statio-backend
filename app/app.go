package app

import (
	"statio/config"
	"statio/container"
	"statio/internal/handlers"
	"statio/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewFiberApp(
	AppConfig *config.AppConfig,
	TableHandler *handlers.TableHandler,
	IndicatorHandler *handlers.IndicatorHandler,
	DimensionHandler *handlers.DimensionHandler,
	OrganizationHandler *handlers.OrganizationHandler,
) (*container.AppContainer, error) {
	App := fiber.New(
		fiber.Config{
			AppName: AppConfig.AppName,
		},
	)

	App.Use(cors.New())
	api := App.Group("/api")
	apiV1 := api.Group("/v1")

	routes.RegisterTableRoutes(apiV1, TableHandler)
	routes.RegisterIndicatorRoutes(apiV1, IndicatorHandler)
	routes.RegisterDimensionRoutes(apiV1, DimensionHandler)
	routes.RegisterOrganizationRoutes(apiV1, OrganizationHandler)

	return &container.AppContainer{
		App:    App,
		Config: AppConfig,
	}, nil
}
