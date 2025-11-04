package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	scheduler, err := InitializeScheduler()

	if err != nil {
		log.Fatalf("failed to initialize scheduler: %v", err)
	}

	log.Println("Scheduler is running...")
	scheduler.Scheduler.Start()

	// Buat channel untuk menangkap sinyal OS
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // Tunggu sinyal
	log.Println("Shutting down scheduler...")

	// Gracefully stop scheduler
	scheduler.Scheduler.Shutdown()

	log.Println("Scheduler stopped. Exiting...")
}
