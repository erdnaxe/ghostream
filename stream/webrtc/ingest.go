package webrtc

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os/exec"

	"github.com/pion/rtp"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
)

func ingestFrom(inputChannel chan srt.Packet) {
	// FIXME Clean code
	var ffmpeg *exec.Cmd
	var ffmpegInput io.WriteCloser
	for {
		var err error = nil
		packet := <-inputChannel
		switch packet.PacketType {
		case "register":
			log.Printf("WebRTC RegisterStream %s", packet.StreamName)

			// From https://github.com/pion/webrtc/blob/master/examples/rtp-to-webrtc/main.go

			// Open a UDP Listener for RTP Packets on port 5004
			videoListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
			if err != nil {
				panic(err)
			}
			audioListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5005})
			if err != nil {
				panic(err)
			}
			defer func() {
				if err = videoListener.Close(); err != nil {
					panic(err)
				}
				if err = audioListener.Close(); err != nil {
					panic(err)
				}
			}()

			ffmpeg = exec.Command("ffmpeg", "-re", "-i", "pipe:0",
				"-an", "-vcodec", "libvpx", //"-cpu-used", "5", "-deadline", "1", "-g", "10", "-error-resilient", "1", "-auto-alt-ref", "1",
				"-f", "rtp", "rtp://127.0.0.1:5004",
				"-vn", "-acodec", "libopus", //"-cpu-used", "5", "-deadline", "1", "-g", "10", "-error-resilient", "1", "-auto-alt-ref", "1",
				"-f", "rtp", "rtp://127.0.0.1:5005")

			fmt.Println("Waiting for RTP Packets, please run GStreamer or ffmpeg now")

			input, err := ffmpeg.StdinPipe()
			if err != nil {
				panic(err)
			}
			ffmpegInput = input
			errOutput, err := ffmpeg.StderrPipe()
			if err != nil {
				panic(err)
			}

			if err := ffmpeg.Start(); err != nil {
				panic(err)
			}

			// Receive video
			go func() {
				for {
					inboundRTPPacket := make([]byte, 1500) // UDP MTU
					n, _, err := videoListener.ReadFromUDP(inboundRTPPacket)
					if err != nil {
						panic(err)
					}
					packet := &rtp.Packet{}
					if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
						panic(err)
					}
					log.Printf("[Video] %s", packet)

					// Write RTP packet to all video tracks
					// Adapt payload and SSRC to match destination
					for _, videoTrack := range videoTracks {
						packet.Header.PayloadType = videoTrack.PayloadType()
						packet.Header.SSRC = videoTrack.SSRC()
						if writeErr := videoTrack.WriteRTP(packet); writeErr != nil {
							panic(err)
						}
					}
				}
			}()

			// Receive audio
			go func() {
				for {
					inboundRTPPacket := make([]byte, 1500) // UDP MTU
					n, _, err := audioListener.ReadFromUDP(inboundRTPPacket)
					if err != nil {
						panic(err)
					}
					packet := &rtp.Packet{}
					if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
						panic(err)
					}
					log.Printf("[Audio] %s", packet)
					for _, audioTrack := range audioTracks {
						packet.Header.PayloadType = audioTrack.PayloadType()
						packet.Header.SSRC = audioTrack.SSRC()
						if writeErr := audioTrack.WriteRTP(packet); writeErr != nil {
							panic(err)
						}
					}
				}
			}()

			go func() {
				scanner := bufio.NewScanner(errOutput)
				for scanner.Scan() {
					log.Printf("[WEBRTC FFMPEG %s] %s", "demo", scanner.Text())
				}
			}()
			break
		case "sendData":
			// FIXME send to stream packet.StreamName
			_, err := ffmpegInput.Write(packet.Data)
			if err != nil {
				panic(err)
			}
			break
		case "close":
			log.Printf("WebRTC CloseConnection %s", packet.StreamName)
			break
		default:
			log.Println("Unknown SRT packet type:", packet.PacketType)
			break
		}
		if err != nil {
			log.Printf("Error occured while receiving SRT packet of type %s: %s", packet.PacketType, err)
		}
	}
}
