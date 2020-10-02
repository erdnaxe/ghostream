package srt

// #include <srt/srt.h>
import "C"

import (
	"gitlab.crans.org/nounous/ghostream/auth"
	"gitlab.crans.org/nounous/ghostream/auth/bypass"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/haivision/srtgo"
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
	if authBackend == nil {
		authBackend, _ = bypass.New()
	}

	options := make(map[string]string)
	options["transtype"] = "live"

	// Start SRT in listen mode
	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	host, port := splitHostPort(cfg.ListenAddress)
	sck := srtgo.NewSrtSocket(host, port, options)
	if err := sck.Listen(cfg.MaxClients); err != nil {
		log.Fatal("Unable to listen to SRT clients:", err)
	}

	// FIXME: See srtgo.SocketOptions and value, err := s.GetSockOptString to get parameters
	// http://ffmpeg.org/ffmpeg-protocols.html#srt

	for {
		// Wait for new connection
		s, err := sck.Accept()
		if err != nil {
			log.Println("Error occurred while accepting request:", err)
			break // FIXME: should not break here
		}

		streamId, err := s.GetSockOptString(C.SRTO_STREAMID)
		if err != nil {
			log.Println("Error while fetching stream key:", err)
			s.Close()
			continue
		}
		if !strings.Contains(streamId, "|") {
			log.Printf("Warning: stream id must be at the format streamId|password. Input: %s", streamId)
			s.Close()
			continue
		}

		splittedStreamId := strings.SplitN(streamId, "|", 2)
		streamName, password := splittedStreamId[0], splittedStreamId[1]
		loggedIn, err := authBackend.Login(streamName, password)
		if !loggedIn {
			log.Printf("Invalid credentials for stream %s.", streamName)
			s.Close()
			continue
		}

		log.Printf("Starting stream %s...", streamName)

		// Create a new buffer
		buff := make([]byte, 2048)

		// Setup stream forwarding
		forwardingChannel <- Packet{StreamName: streamName, PacketType: "register", Data: nil}

		// Read RTP packets forever and send them to the WebRTC Client
		for {
			n, err := s.Read(buff, 10000)
			if err != nil {
				log.Println("Error occured while reading SRT socket:", err)
				break
			}

			if n == 0 {
				// End of stream
				log.Printf("Received no bytes, stopping stream.")
				break
			}
			// log.Printf("Received %d bytes", n)

			// Send raw packet to other streams
			// Copy data in another buffer to ensure that the data would not be overwritten
			data := make([]byte, n)
			copy(data, buff[:n])
			forwardingChannel <- Packet{StreamName: streamName, PacketType: "sendData", Data: data}

			// TODO: Send to WebRTC
			// See https://github.com/ebml-go/webm/blob/master/reader.go
			//err := videoTrack.WriteSample(media.Sample{Data: data, Samples: uint32(sampleCount)})
		}

		forwardingChannel <- Packet{StreamName: streamName, PacketType: "close", Data: nil}
	}

	sck.Close()
}
