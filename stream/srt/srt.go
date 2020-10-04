package srt

// #include <srt/srt.h>
import "C"

import (
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/haivision/srtgo"
	"gitlab.crans.org/nounous/ghostream/auth"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
	MaxClients    int
}

// Packet contains the necessary data to broadcast events like stream creating, packet receiving or stream closing.
type Packet struct {
	Data       []byte
	PacketType string
	StreamName string
}

// Split host and port from listen address
func splitHostPort(hostport string) (string, uint16) {
	host, portS, err := net.SplitHostPort(hostport)
	if err != nil {
		log.Fatalf("Failed to split host and port from %s", hostport)
	}
	if host == "" {
		host = "0.0.0.0"
	}
	port64, err := strconv.ParseUint(portS, 10, 16)
	if err != nil {
		log.Fatalf("Port is not a integer: %s", err)
	}
	return host, uint16(port64)
}

// Serve SRT server
func Serve(cfg *Options, authBackend auth.Backend, forwardingChannel chan Packet) {
	// Start SRT in listening mode
	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	host, port := splitHostPort(cfg.ListenAddress)
	sck := srtgo.NewSrtSocket(host, port, nil)
	if err := sck.Listen(cfg.MaxClients); err != nil {
		log.Fatal("Unable to listen for SRT clients:", err)
	}

	clientDataChannels := make([]chan Packet, 0, cfg.MaxClients)

	for {
		// Wait for new connection
		s, err := sck.Accept()
		if err != nil {
			// Something wrong happenned
			continue
		}

		// streamid can be "name:password" for streamer or "name" for viewer
		streamID, err := s.GetSockOptString(C.SRTO_STREAMID)
		if err != nil {
			log.Print("Failed to get socket streamid")
			continue
		}
		split := strings.Split(streamID, ":")

		if len(split) > 1 {
			// password was provided so it is a streamer
			name, password := split[0], split[1]
			if authBackend != nil {
				// check password
				if ok, err := authBackend.Login(name, password); !ok || err != nil {
					log.Printf("Failed to authenticate for stream %s", name)
					s.Close()
					continue
				}
			}

			go handleStreamer(s, name, &clientDataChannels, forwardingChannel)
		} else {
			// password was not provided so it is a viewer
			name := split[0]

			dataChannel := make(chan Packet, 4096)
			clientDataChannels = append(clientDataChannels, dataChannel)

			go handleViewer(s, name, dataChannel, &clientDataChannels)
		}
	}
}
