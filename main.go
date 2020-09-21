package main

import (
	"log"

	"gitlab.crans.org/nounous/ghostream/internal/config"
	"gitlab.crans.org/nounous/ghostream/web"
)

func main() {
	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Start web server
	web.ServeHTTP(cfg)
}
