package main

import (
	"fmt"

	"log"
)

func main() {
	app, err := InitializeApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	log.Fatal(app.App.Listen(fmt.Sprintf(":%s", app.Config.AppPort)))
}
