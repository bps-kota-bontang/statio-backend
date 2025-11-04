package config

import (
	"os"
)

// RedisConfig menyimpan konfigurasi Redis
type RedisConfig struct {
	RedisHost string
	RedisPort string
}

// LoadRedisConfig membaca konfigurasi dari environment variable
func LoadRedisConfig() (*RedisConfig, error) {
	return &RedisConfig{
		RedisHost: os.Getenv("REDIS_HOST"),
		RedisPort: os.Getenv("REDIS_PORT"),
	}, nil
}
