package srt

import (
	"log"

	"github.com/haivision/srtgo"
	"github.com/pion/rtp"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
}

// Serve SRT server
func Serve(cfg *Options) {
	log.Printf("SRT server listening on %s", cfg.ListenAddress)

	options := make(map[string]string)
	options["transtype"] = "live"

	// FIXME: cfg.ListenAddress -> host and port
	sck := srtgo.NewSrtSocket("0.0.0.0", 9710, options)
	sck.Listen(1)

	for {
		s, err := sck.Accept()
		if err != nil {
			log.Println("Error occurred while accepting request:", err)
			continue
		}

		buff := make([]byte, 2048)
		n, err := s.Read(buff, 10000)
		if err != nil {
			log.Println("Error occurred while reading SRT socket:", err)
			break
		}
		if n == 0 {
			// End of stream
			break
		}

		// Unmarshal the incoming packet
		packet := &rtp.Packet{}
		if err = packet.Unmarshal(buff[:n]); err != nil {
			log.Println("Error occured while unmarshaling SRT:", err)
			break
		}

		// videoTrack, err := peerConnection.NewTrack(payloadType, packet.SSRC, "video", "pion")

		// Read RTP packets forever and send them to the WebRTC Client
		for {
			n, err := s.Read(buff, 10000)
			if err != nil {
				log.Println("Error occured while reading SRT socket:", err)
				break
			}

			log.Printf("Received %d bytes", n)

			packet := &rtp.Packet{}
			if err := packet.Unmarshal(buff[:n]); err != nil {
				panic(err)
			}
			payloadType := uint8(22) // FIXME put vp8 payload
			packet.Header.PayloadType = payloadType

			//err := videoTrack.WriteRTP(packet)
		}
	}
}
