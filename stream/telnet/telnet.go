// Package telnet provides some fancy tools, like an ASCII-art stream.
package telnet

import (
	"log"
	"net"
	"time"

	"gitlab.crans.org/nounous/ghostream/stream"
)

// Options holds telnet package configuration
type Options struct {
	Enabled       bool
	ListenAddress string
	Width         int
	Height        int
	Delay         int
}

// Serve Telnet server
func Serve(streams map[string]*stream.Stream, cfg *Options) {
	if !cfg.Enabled {
		// Telnet is not enabled, ignore
		return
	}

	// Start conversion routine
	textStreams := make(map[string]*[]byte)
	go autoStartConversion(streams, textStreams, cfg)

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

		go handleViewer(s, streams, textStreams, cfg)
	}
}

// Convertion routine listen to existing stream and start text conversions
func autoStartConversion(streams map[string]*stream.Stream, textStreams map[string]*[]byte, cfg *Options) {
	for {
		for name, stream := range streams {
			textStream, ok := textStreams[name]
			if ok {
				// Everything is fine
				continue
			}

			// Start conversion
			log.Print("Starting text conversion of %s", name)
			textStream = &[]byte{}
			textStreams[name] = textStream
			go streamToTextStream(stream, textStream, cfg)
		}
		time.Sleep(time.Second)
	}
}
