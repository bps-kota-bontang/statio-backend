package providers

import (
	"statio/config"
	"statio/internal/services"
	"time"
)

func NewJWTProvider(
	cfg *config.AuthConfig,
) *services.JWTService {
	return services.NewJWTService(
		cfg.AuthJWTSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
}
