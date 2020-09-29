package srt

import (
	"gitlab.crans.org/nounous/ghostream/stream/multicast"
	"log"
	"net"
	"strconv"

	"github.com/haivision/srtgo"
	"github.com/pion/rtp"
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

	log.Printf("SRT server listening on %s", cfg.ListenAddress)
	host, port := splitHostPort(cfg.ListenAddress)
	sck := srtgo.NewSrtSocket(host, uint16(port), options)
	sck.Listen(cfg.MaxClients)

	for {
		s, err := sck.Accept()
		if err != nil {
			log.Println("Error occurred while accepting request:", err)
			continue
		}

		multicast.RegisterStream("demo") // FIXME Replace with real stream key

		buff := make([]byte, 2048)
		n, err := s.Read(buff, 10000)
		if err != nil {
			log.Println("Error occurred while reading SRT socket:", err)
			break
		}
		if n == 0 {
			// End of stream
			multicast.CloseConnection("demo")
			break
		}

		multicast.SendPacket("demo", buff[:n])

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
				multicast.CloseConnection("demo")
				break
			}

			multicast.SendPacket("demo", buff[:n])

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
