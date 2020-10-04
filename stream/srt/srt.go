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

	// FIXME Better structure
	clientDataChannels := make([]chan Packet, cfg.MaxClients)
	listeners := 0

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

			go handleStreamer(s, name, clientDataChannels, &listeners, forwardingChannel)
		} else {
			// password was not provided so it is a viewer
			name := split[0]

			dataChannel := make(chan Packet, 2048)
			clientDataChannels[listeners] = dataChannel
			listeners++

			go handleViewer(s, name, dataChannel)
		}
	}
}

func handleStreamer(s *srtgo.SrtSocket, name string, clientDataChannels []chan Packet, listeners *int, forwardingChannel chan Packet) {
	log.Printf("New SRT streamer for stream %s", name)

	// Create a new buffer
	buff := make([]byte, 2048)

	// Setup stream forwarding
	forwardingChannel <- Packet{StreamName: name, PacketType: "register", Data: nil}

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
		forwardingChannel <- Packet{StreamName: name, PacketType: "sendData", Data: data}
		for i := 0; i < *listeners; i++ {
			clientDataChannels[i] <- Packet{StreamName: name, PacketType: "sendData", Data: data}
		}

		// TODO: Send to WebRTC
		// See https://github.com/ebml-go/webm/blob/master/reader.go
		//err := videoTrack.WriteSample(media.Sample{Data: data, Samples: uint32(sampleCount)})
	}

	forwardingChannel <- Packet{StreamName: name, PacketType: "close", Data: nil}
}

func handleViewer(s *srtgo.SrtSocket, name string, dataChannel chan Packet) {
	log.Printf("New SRT viewer for stream %s", name)

	// Receive packets from channel and send them
	for {
		packet := <-dataChannel
		if packet.PacketType == "sendData" {
			_, err := s.Write(packet.Data, 10000)
			if err != nil {
				s.Close()
				return
			}
		}
	}
}
