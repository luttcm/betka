package main

import (
	"log"

	"bet/backend/internal/config"
	"bet/backend/internal/http"
)

func main() {
	cfg := config.Load()
	router := http.NewRouter()

	log.Printf("API started on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run API: %v", err)
	}
}
