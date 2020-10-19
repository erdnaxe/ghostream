// Package srt serves a SRT server
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
	"gitlab.crans.org/nounous/ghostream/messaging"
)

// Options holds web package configuration
type Options struct {
	Enabled       bool
	ListenAddress string
	MaxClients    int
}

// Split host and port from listen address
func splitHostPort(hostport string) (string, uint16, error) {
	host, portS, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	}
	if host == "" {
		host = "0.0.0.0"
	}
	port64, err := strconv.ParseUint(portS, 10, 16)
	if err != nil {
		return "", 0, err
	}
	return host, uint16(port64), nil
}

// Serve SRT server
func Serve(streams *messaging.Streams, authBackend auth.Backend, cfg *Options) {
	if !cfg.Enabled {
		// SRT is not enabled, ignore
		return
	}

	// Start SRT in listening mode
	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	host, port, err := splitHostPort(cfg.ListenAddress)
	if err != nil {
		log.Fatalf("Failed to split host and port from %s", cfg.ListenAddress)
	}

	options := make(map[string]string)
	options["blocking"] = "0"
	options["transtype"] = "live"
	sck := srtgo.NewSrtSocket(host, port, options)
	if err := sck.Listen(cfg.MaxClients); err != nil {
		log.Fatal("Unable to listen for SRT clients:", err)
	}

	for {
		// Wait for new connection
		s, err := sck.Accept()
		if err != nil {
			// Something wrong happened
			log.Println(err)
			continue
		}

		// FIXME: Flush socket
		// Without this, the SRT buffer might get full before reading it

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

			go handleStreamer(s, streams, name)
		} else {
			// password was not provided so it is a viewer
			name := split[0]

			// Send stream
			go handleViewer(s, streams, name)
		}
	}
}
