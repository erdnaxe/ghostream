// Package webrtc provides the backend to simulate a WebRTC client to send stream
package webrtc

import (
	"bufio"
	"fmt"
	"gitlab.crans.org/nounous/ghostream/stream/telnet"
	"io"
	"log"
	"net"
	"os/exec"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
)

func ingestFrom(inputChannel chan srt.Packet) {
	// FIXME Clean code
	var ffmpeg *exec.Cmd
	var ffmpegInput io.WriteCloser

	for {
		var err error = nil
		srtPacket := <-inputChannel
		switch srtPacket.PacketType {
		case "register":
			log.Printf("WebRTC RegisterStream %s", srtPacket.StreamName)

			// Open a UDP Listener for RTP Packets on port 5004
			videoListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
			if err != nil {
				log.Printf("Faited to open UDP listener %s", err)
				return
			}
			audioListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5005})
			if err != nil {
				log.Printf("Faited to open UDP listener %s", err)
				return
			}
			defer func() {
				if err = videoListener.Close(); err != nil {
					log.Printf("Faited to close UDP listener %s", err)
				}
				if err = audioListener.Close(); err != nil {
					log.Printf("Faited to close UDP listener %s", err)
				}
			}()

			ffmpeg = exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-re", "-i", "pipe:0",
				"-an", "-vcodec", "libvpx", "-crf", "10", "-cpu-used", "5", "-b:v", "6000k", "-maxrate", "8000k", "-bufsize", "12000k", // TODO Change bitrate when changing quality
				"-qmin", "10", "-qmax", "42", "-threads", "4", "-deadline", "1", "-error-resilient", "1",
				"-auto-alt-ref", "1",
				"-f", "rtp", "rtp://127.0.0.1:5004",
				"-vn", "-acodec", "libopus", "-cpu-used", "5", "-deadline", "1", "-qmin", "10", "-qmax", "42", "-error-resilient", "1", "-auto-alt-ref", "1",
				"-f", "rtp", "rtp://127.0.0.1:5005",
				"-an", "-f", "rawvideo", "-vf", fmt.Sprintf("scale=%dx%d", telnet.Cfg.Width, telnet.Cfg.Height), "-pix_fmt", "gray", "pipe:1")

			input, err := ffmpeg.StdinPipe()
			if err != nil {
				panic(err)
			}
			ffmpegInput = input
			errOutput, err := ffmpeg.StderrPipe()
			if err != nil {
				panic(err)
			}
			output, err := ffmpeg.StdoutPipe()
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
						log.Printf("Failed to read from UDP: %s", err)
						continue
					}
					packet := &rtp.Packet{}
					if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
						log.Printf("Failed to unmarshal RTP srtPacket: %s", err)
						continue
					}

					if videoTracks[srtPacket.StreamName] == nil {
						videoTracks[srtPacket.StreamName] = make([]*webrtc.Track, 0)
					}

					// Write RTP srtPacket to all video tracks
					// Adapt payload and SSRC to match destination
					for _, videoTrack := range videoTracks[srtPacket.StreamName] {
						packet.Header.PayloadType = videoTrack.PayloadType()
						packet.Header.SSRC = videoTrack.SSRC()
						if writeErr := videoTrack.WriteRTP(packet); writeErr != nil {
							log.Printf("Failed to write to video track: %s", err)
							continue
						}
					}
				}
			}()

			// Receive ascii
			go telnet.ServeAsciiArt(output)

			// Receive audio
			go func() {
				for {
					inboundRTPPacket := make([]byte, 1500) // UDP MTU
					n, _, err := audioListener.ReadFromUDP(inboundRTPPacket)
					if err != nil {
						log.Printf("Failed to read from UDP: %s", err)
						continue
					}
					packet := &rtp.Packet{}
					if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
						log.Printf("Failed to unmarshal RTP srtPacket: %s", err)
						continue
					}

					if audioTracks[srtPacket.StreamName] == nil {
						audioTracks[srtPacket.StreamName] = make([]*webrtc.Track, 0)
					}

					// Write RTP srtPacket to all audio tracks
					// Adapt payload and SSRC to match destination
					for _, audioTrack := range audioTracks[srtPacket.StreamName] {
						packet.Header.PayloadType = audioTrack.PayloadType()
						packet.Header.SSRC = audioTrack.SSRC()
						if writeErr := audioTrack.WriteRTP(packet); writeErr != nil {
							log.Printf("Failed to write to audio track: %s", err)
							continue
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
			// FIXME send to stream srtPacket.StreamName
			if _, err := ffmpegInput.Write(srtPacket.Data); err != nil {
				log.Printf("Failed to write data to ffmpeg input: %s", err)
			}
			break
		case "close":
			log.Printf("WebRTC CloseConnection %s", srtPacket.StreamName)
			break
		default:
			log.Println("Unknown SRT srtPacket type:", srtPacket.PacketType)
			break
		}
		if err != nil {
			log.Printf("Error occured while receiving SRT srtPacket of type %s: %s", srtPacket.PacketType, err)
		}
	}
}
