package config

import "os"

type AuthConfig struct {
	AuthGateURL   string
	AuthJWTSecret string
	AuthGateID    string
}

func LoadAuthConfig() (*AuthConfig, error) {
	return &AuthConfig{
		AuthGateURL:   os.Getenv("AUTH_GATE_URL"),
		AuthJWTSecret: os.Getenv("AUTH_JWT_SECRET"),
		AuthGateID:    os.Getenv("AUTH_GATE_ID"),
	}, nil
}
