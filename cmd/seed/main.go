package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"statio/config"
	"statio/internal/providers"
)

func main() {
	dummy := flag.Bool("dummy", false, "seed dummy projects, tables, and indicators")
	flag.Parse()

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
	if *dummy {
		if err := providers.SeedDummyData(db); err != nil {
			log.Fatalf("failed to seed dummy data: %v", err)
		}
		log.Println("seeded dummy projects, tables, indicators, and project links")
		fmt.Println("Admin and dummy data seed completed successfully")
		return
	}
	fmt.Println("Admin seed completed successfully")
}
