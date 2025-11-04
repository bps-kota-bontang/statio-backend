package container

import (
	"statio/config"

	"github.com/gofiber/fiber/v2"
)

type AppContainer struct {
	App    *fiber.App
	Config *config.AppConfig
}
