// Package telnet expose text version of stream.
package telnet

import (
	"log"
	"net"
	"strings"
	"time"

	"gitlab.crans.org/nounous/ghostream/messaging"
)

// Options holds telnet package configuration
type Options struct {
	Enabled       bool
	ListenAddress string
}

// Serve Telnet server
func Serve(streams *messaging.Streams, cfg *Options) {
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
		socket, err := listener.Accept()
		if err != nil {
			log.Printf("Error while accepting TCP socket: %s", err)
			continue
		}

		go handleViewer(socket, streams, cfg)
	}
}

func handleViewer(s net.Conn, streams *messaging.Streams, cfg *Options) {
	// Prompt user about stream name
	if _, err := s.Write([]byte("[GHOSTREAM]\nEnter stream name: ")); err != nil {
		log.Printf("Error while writing to TCP socket: %s", err)
		s.Close()
		return
	}
	buff := make([]byte, 255)
	n, err := s.Read(buff)
	if err != nil {
		log.Printf("Error while requesting stream ID to telnet client: %s", err)
		s.Close()
		return
	}
	name := strings.TrimSpace(string(buff[:n]))
	if len(name) < 1 {
		// Too short, exit
		s.Close()
		return
	}

	// Wait a bit
	time.Sleep(time.Second)

	// Get requested stream
	stream, err := streams.Get(name)
	if err != nil {
		log.Printf("Kicking new Telnet viewer: %s", err)
		if _, err := s.Write([]byte("This stream is inactive.\n")); err != nil {
			log.Printf("Error while writing to TCP socket: %s", err)
		}
		s.Close()
		return
	}

	// Get requested quality
	qualityName := "text"
	q, err := stream.GetQuality(qualityName)
	if err != nil {
		log.Printf("Kicking new Telnet viewer: %s", err)
		if _, err := s.Write([]byte("This stream is not converted to text.\n")); err != nil {
			log.Printf("Error while writing to TCP socket: %s", err)
		}
		s.Close()
		return
	}
	log.Printf("New Telnet viewer for stream %s quality %s", name, qualityName)

	// Register new client
	c := make(chan []byte, 128)
	q.Register(c)
	stream.IncrementClientCount()

	// Hide terminal cursor
	if _, err = s.Write([]byte("\033[?25l")); err != nil {
		log.Printf("Error while writing to TCP socket: %s", err)
		s.Close()
		return
	}

	// Receive data and send them
	for data := range c {
		if len(data) < 1 {
			log.Print("Remove Telnet viewer because of end of stream")
			break
		}

		// Send data
		_, err := s.Write(data)
		if err != nil {
			log.Printf("Remove Telnet viewer because of sending error, %s", err)
			break
		}
	}

	// Close output
	q.Unregister(c)
	stream.DecrementClientCount()
	s.Close()
}
