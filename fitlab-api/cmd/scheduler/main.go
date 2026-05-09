package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fitlab-api/internal/database"
	"fitlab-api/internal/services"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./fitlab.db"
	}

	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Scheduler started, database: %s", dbPath)

	intervalStr := os.Getenv("CHECK_INTERVAL")
	interval := 5 * time.Minute
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil && d > 0 {
			interval = d
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		services.StartNotifierLoop(db.DB, interval)
	}()

	<-sigChan
	log.Println("Scheduler shutting down")
}
