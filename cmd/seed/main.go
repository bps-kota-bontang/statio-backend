package main

import (
	"fmt"
	"log"
	"os"
	"statio/config"
	"statio/internal/providers"
)

func main() {
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		log.Fatalf("failed to load app config: %v", err)
	}

	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("failed to load database config: %v", err)
	}

	db, err := providers.NewDBConnection(dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	username := os.Getenv("SEED_ADMIN_USERNAME")
	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")

	user, err := providers.SeedDataAdmin(db, username, email, password)
	if err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	log.Printf("seeded admin user: %s (%s)", user.Username, appConfig.AppName)
	fmt.Println("Admin seed completed successfully")
}
