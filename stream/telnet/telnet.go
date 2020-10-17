// Package telnet provides some fancy tools, like an ASCII-art stream.
package telnet

import (
	"log"
	"net"

	"gitlab.crans.org/nounous/ghostream/stream"
)

// Options holds telnet package configuration
type Options struct {
	Enabled       bool
	ListenAddress string
}

// Serve Telnet server
func Serve(streams map[string]*stream.Stream, cfg *Options) {
	if !cfg.Enabled {
		// Telnet is not enabled, ignore
		return
	}

	// Start TCP server
	listener, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		log.Fatalf("Error while listening to the address %s: %s", cfg.ListenAddress, err)
	}
	log.Printf("Telnet server listening on %s", cfg.ListenAddress)

	// Handle each new client
	for {
		s, err := listener.Accept()
		if err != nil {
			log.Printf("Error while accepting TCP socket: %s", s)
			continue
		}

		go handleViewer(s, streams, cfg)
	}
}
