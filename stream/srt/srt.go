package srt

// #include <srt/srt.h>
import "C"

import (
	"fmt"
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

var (
	authBackend       auth.Backend
	forwardingChannel chan Packet
)

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
func Serve(cfg *Options, backend auth.Backend, forwarding chan Packet) {
	if backend == nil {
		backend, _ = bypass.New()
	}
	authBackend = backend
	forwardingChannel = forwarding

	options := make(map[string]string)
	options["transtype"] = "file"
	options["mode"] = "listener"

	// Start SRT in listen mode
	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	host, port := splitHostPort(cfg.ListenAddress)
	sck := srtgo.NewSrtSocket(host, port, options)
	if err := sck.Listen(cfg.MaxClients); err != nil {
		log.Fatal("Unable to listen to SRT clients:", err)
	}

	// FIXME: See srtgo.SocketOptions and value, err := s.GetSockOptString to get parameters
	// http://ffmpeg.org/ffmpeg-protocols.html#srt

	// FIXME: Get the stream type
	streamStarted := false
	// FIXME Better structure
	clientDataChannels := make([]chan Packet, cfg.MaxClients)
	listeners := 0

	for {
		// Wait for new connection
		s, err := sck.Accept()
		if err != nil {
			// log.Println("Error occurred while accepting request:", err)
			continue // break // FIXME: should not break here
		}

		if !streamStarted {
			go acceptCallerSocket(s, clientDataChannels, &listeners)
			streamStarted = true
		} else {
			dataChannel := make(chan Packet, 2048)
			clientDataChannels[listeners] = dataChannel
			listeners++
			go acceptListeningSocket(s, dataChannel)
		}
	}

	sck.Close()
}

func acceptCallerSocket(s *srtgo.SrtSocket, clientDataChannels []chan Packet, listeners *int) {
	streamName, err := authenticateSocket(s)
	if err != nil {
		log.Println("Authentication failure:", err)
		s.Close()
		return
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
		for i := 0; i < *listeners; i++ {
			clientDataChannels[i] <- Packet{StreamName: streamName, PacketType: "sendData", Data: data}
		}

		// TODO: Send to WebRTC
		// See https://github.com/ebml-go/webm/blob/master/reader.go
		//err := videoTrack.WriteSample(media.Sample{Data: data, Samples: uint32(sampleCount)})
	}

	forwardingChannel <- Packet{StreamName: streamName, PacketType: "close", Data: nil}
}

func acceptListeningSocket(s *srtgo.SrtSocket, dataChannel chan Packet) {
	streamName, err := s.GetSockOptString(C.SRTO_STREAMID)
	if err != nil {
		panic(err)
	}
	log.Printf("New listener for stream %s", streamName)

	for {
		packet := <-dataChannel
		if packet.PacketType == "sendData" {
			_, err := s.Write(packet.Data, 10000)
			if err != nil {
				s.Close()
				break
			}
		}
	}
}

func authenticateSocket(s *srtgo.SrtSocket) (string, error) {
	streamID, err := s.GetSockOptString(C.SRTO_STREAMID)
	if err != nil {
		return "", fmt.Errorf("error while fetching stream key: %s", err)
	}
	log.Println(s.GetSockOptString(C.SRTO_PASSPHRASE))
	if !strings.Contains(streamID, ":") {
		return streamID, fmt.Errorf("warning: stream id must be at the format streamID:password. Input: %s", streamID)
	}

	splittedStreamID := strings.SplitN(streamID, ":", 2)
	streamName, password := splittedStreamID[0], splittedStreamID[1]
	loggedIn, err := authBackend.Login(streamName, password)
	if !loggedIn {
		return streamID, fmt.Errorf("invalid credentials for stream %s", streamName)
	}

	return streamName, nil
}
