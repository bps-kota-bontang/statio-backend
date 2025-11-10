package providers

import (
	"statio/config"
	"statio/internal/services"
	"time"
)

func NewJWTProvider(cfg *config.AppConfig) *services.JWTService {
	return services.NewJWTService(
		cfg.AppJWTSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
}
