package main

import (
	"log"

	"gitlab.crans.org/nounous/ghostream/internal/config"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/web"
)

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Start web server routine
	go func() {
		web.ServeHTTP(cfg)
	}()

	// Start monitoring server routine
	go func() {
		monitoring.ServeHTTP(cfg)
	}()

	// Wait for routines
	select {}
}
