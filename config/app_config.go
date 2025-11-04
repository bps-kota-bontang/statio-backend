package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig menyimpan konfigurasi aplikasi
type AppConfig struct {
	AppName string
	AppEnv  string
	AppPort string
}

// LoadAppConfig membaca variabel lingkungan dan mengembalikan pointer ke AppConfig
func LoadAppConfig() (*AppConfig, error) {
	// Cek env mode (harus dilakukan sebelum load .env)
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	// Hanya load .env jika bukan production
	if appEnv != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Error loading .env file: %v", err)
		} else {
			log.Println("Loaded .env file")
		}
	}

	appName := os.Getenv("APP_NAME")
	appPort := os.Getenv("APP_PORT")

	if appPort == "" {
		appPort = "3000" // default port
	}

	return &AppConfig{
		AppName: appName,
		AppEnv:  appEnv,
		AppPort: appPort,
	}, nil
}
