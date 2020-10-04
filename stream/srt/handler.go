package srt

import (
	"log"

	"github.com/haivision/srtgo"
)

func handleStreamer(s *srtgo.SrtSocket, name string, clientDataChannels *[]chan Packet, forwardingChannel chan Packet) {
	log.Printf("New SRT streamer for stream %s", name)

	// Create a new buffer
	buff := make([]byte, 2048)

	// Setup stream forwarding
	forwardingChannel <- Packet{StreamName: name, PacketType: "register", Data: nil}

	// Read RTP packets forever and send them to the WebRTC Client
	for {
		// UDP packet cannot be larger than MTU (1500)
		n, err := s.Read(buff, 1500)
		if err != nil {
			log.Println("Error occured while reading SRT socket:", err)
			break
		}

		if n == 0 {
			// End of stream
			log.Printf("Received no bytes, stopping stream.")
			break
		}

		// Send raw packet to other streams
		// Copy data in another buffer to ensure that the data would not be overwritten
		data := make([]byte, n)
		copy(data, buff[:n])
		forwardingChannel <- Packet{StreamName: name, PacketType: "sendData", Data: data}
		for _, dataChannel := range *clientDataChannels {
			dataChannel <- Packet{StreamName: name, PacketType: "sendData", Data: data}
		}
	}

	forwardingChannel <- Packet{StreamName: name, PacketType: "close", Data: nil}
}

func handleViewer(s *srtgo.SrtSocket, name string, dataChannel chan Packet, dataChannels *[]chan Packet) {
	// FIXME Should not pass all dataChannels to one viewer

	log.Printf("New SRT viewer for stream %s", name)

	// Receive packets from channel and send them
	for {
		packet := <-dataChannel
		if packet.PacketType == "sendData" {
			_, err := s.Write(packet.Data, 10000)
			if err != nil {
				s.Close()
				for i, channel := range *dataChannels {
					if channel == dataChannel {
						*dataChannels = append((*dataChannels)[:i], (*dataChannels)[i+1:]...)
					}
				}
				return
			}
		}
	}
}
