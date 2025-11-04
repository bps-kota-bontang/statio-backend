package main

import (
	"fmt"

	"log"
)

func main() {
	worker, err := InitializeWorker()
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}

	fmt.Println("Worker is running...")

	if err := worker.Server.Run(worker.Mux); err != nil {
		log.Fatalf("failed to start worker: %v", err)
	}
}
