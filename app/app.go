package app

import (
	"statio/config"
	"statio/container"
	"statio/internal/handlers"
	"statio/internal/middlewares"
	"statio/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewFiberApp(
	AppConfig *config.AppConfig,
	JWTMiddleware *middlewares.JWTMiddleware,
	AuthHandler *handlers.AuthHandler,
	TableHandler *handlers.TableHandler,
	IndicatorHandler *handlers.IndicatorHandler,
	DimensionHandler *handlers.DimensionHandler,
	OrganizationHandler *handlers.OrganizationHandler,
	UserHandler *handlers.UserHandler,
	DashboardHandler *handlers.DashboardHandler,
	IntegrationHandler *handlers.IntegrationHandler,
	ConfigurationHandler *handlers.ConfigurationHandler,
) (*container.AppContainer, error) {
	App := fiber.New(
		fiber.Config{
			AppName: AppConfig.AppName,
		},
	)

	isProd := AppConfig.AppEnv == "production"

	corsConfig := cors.Config{
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Authorization, X-Refresh-Attempt",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Disposition",
	}

	if isProd {
		corsConfig.AllowOrigins = AppConfig.AppURL
	} else {
		corsConfig.AllowOrigins = "http://localhost:5173"
	}

	App.Use(cors.New(corsConfig))

	App.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Statio (Build: " + AppConfig.AppBuild + ")")
	})

	App.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	api := App.Group("/api")
	apiV1 := api.Group("/v1")

	// Public
	routes.RegisterAuthRoutes(apiV1, AuthHandler, JWTMiddleware)

	// Protected
	protected := apiV1.Group("/", JWTMiddleware.Protected())
	routes.RegisterTableRoutes(protected, TableHandler)
	routes.RegisterIndicatorRoutes(protected, IndicatorHandler)
	routes.RegisterDimensionRoutes(protected, DimensionHandler)
	routes.RegisterOrganizationRoutes(protected, OrganizationHandler)
	routes.RegisterUserRoutes(protected, UserHandler)
	routes.RegisterDashboardRoutes(protected, DashboardHandler)
	routes.RegisterIntegrationRoutes(protected, IntegrationHandler)
	routes.RegisterConfigurationRoutes(protected, ConfigurationHandler)

	return &container.AppContainer{
		App:    App,
		Config: AppConfig,
	}, nil
}
