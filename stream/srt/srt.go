package srt

import (
	"log"
	"net"
	"strconv"

	"github.com/haivision/srtgo"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
	MaxClients    int
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
func Serve(cfg *Options) {
	options := make(map[string]string)
	options["transtype"] = "live"

	// Start SRT in listen mode
	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	host, port := splitHostPort(cfg.ListenAddress)
	sck := srtgo.NewSrtSocket(host, uint16(port), options)
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

		// Create a new buffer
		buff := make([]byte, 2048)

		// Setup stream forwarding
		// FIXME: demo should be replaced by stream name
		/*	if err := forwarding.RegisterStream("demo"); err != nil {
			log.Println("Error occurred during forward stream init:", err)
			break
		} */

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
			// forwarding.SendPacket("demo", buff[:n])

			// TODO: Send to WebRTC
			// See https://github.com/ebml-go/webm/blob/master/reader.go
			//err := videoTrack.WriteSample(media.Sample{Data: data, Samples: uint32(sampleCount)})
		}

		/*	if err := forwarding.CloseConnection("demo"); err != nil {
			log.Printf("Failed to close forward stream: %s", err)
		} */
	}
}
