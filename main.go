//go:generate pkger

// Package main provides the full-featured server with configuration loading
// and communication between routines.
package main

import (
	"gitlab.crans.org/nounous/ghostream/stream/ovenmediaengine"
	"log"

	"github.com/pkg/profile"
	"gitlab.crans.org/nounous/ghostream/auth"
	"gitlab.crans.org/nounous/ghostream/internal/config"
	"gitlab.crans.org/nounous/ghostream/internal/monitoring"
	"gitlab.crans.org/nounous/ghostream/messaging"
	"gitlab.crans.org/nounous/ghostream/stream/forwarding"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
	"gitlab.crans.org/nounous/ghostream/stream/telnet"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
	"gitlab.crans.org/nounous/ghostream/transcoder"
	"gitlab.crans.org/nounous/ghostream/web"
)

func main() {
	// TODO Don't always profile if not needed
	defer profile.Start().Stop()

	// Configure logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln("Failed to load configuration:", err)
	}

	// Init authentification
	authBackend, err := auth.New(&cfg.Auth)
	if err != nil {
		log.Fatalln("Failed to load authentification backend:", err)
	}
	if authBackend != nil {
		defer authBackend.Close()
	}

	// Init streams messaging
	streams := messaging.New()

	// Start routines
	go transcoder.Init(streams, &cfg.Transcoder)
	go forwarding.Serve(streams, cfg.Forwarding)
	go monitoring.Serve(&cfg.Monitoring)
	go ovenmediaengine.Serve(streams, &cfg.OME)
	go srt.Serve(streams, authBackend, &cfg.Srt)
	go telnet.Serve(streams, &cfg.Telnet)
	go web.Serve(streams, &cfg.Web, &cfg.OME)
	go webrtc.Serve(streams, &cfg.WebRTC)

	// Wait for routines
	select {}
}
