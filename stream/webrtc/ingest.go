// Package webrtc provides the backend to simulate a WebRTC client to send stream
package webrtc

import (
	"bufio"
	"log"
	"net"
	"os/exec"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"gitlab.crans.org/nounous/ghostream/messaging"
)

func ingest(name string, q *messaging.Quality) {
	// Register to get stream
	videoInput := make(chan []byte, 1024)
	q.Register(videoInput)

	// Open a UDP Listener for RTP Packets on port 5004
	audioListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5004})
	if err != nil {
		log.Printf("Faited to open UDP listener %s", err)
		return
	}
	videoListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5005})
	if err != nil {
		log.Printf("Faited to open UDP listener %s", err)
		return
	}

	// Start ffmpag to convert videoInput to video and audio UDP
	ffmpeg, err := startFFmpeg(videoInput)
	if err != nil {
		log.Printf("Error while starting ffmpeg: %s", err)
		return
	}

	// Receive video
	go func() {
		inboundRTPPacket := make([]byte, 1500) // UDP MTU
		for {
			n, _, err := videoListener.ReadFromUDP(inboundRTPPacket)
			if err != nil {
				log.Printf("Failed to read from UDP: %s", err)
				break
			}
			packet := &rtp.Packet{}
			if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
				log.Printf("Failed to unmarshal RTP srtPacket: %s", err)
				continue
			}

			if videoTracks[name] == nil {
				videoTracks[name] = make([]*webrtc.Track, 0)
			}

			// Write RTP srtPacket to all video tracks
			// Adapt payload and SSRC to match destination
			for _, videoTrack := range videoTracks[name] {
				packet.Header.PayloadType = videoTrack.PayloadType()
				packet.Header.SSRC = videoTrack.SSRC()
				if writeErr := videoTrack.WriteRTP(packet); writeErr != nil {
					log.Printf("Failed to write to video track: %s", err)
					continue
				}
			}
		}
	}()

	// Receive audio
	go func() {
		inboundRTPPacket := make([]byte, 1500) // UDP MTU
		for {
			n, _, err := audioListener.ReadFromUDP(inboundRTPPacket)
			if err != nil {
				log.Printf("Failed to read from UDP: %s", err)
				break
			}
			packet := &rtp.Packet{}
			if err := packet.Unmarshal(inboundRTPPacket[:n]); err != nil {
				log.Printf("Failed to unmarshal RTP srtPacket: %s", err)
				continue
			}

			if audioTracks[name] == nil {
				audioTracks[name] = make([]*webrtc.Track, 0)
			}

			// Write RTP srtPacket to all audio tracks
			// Adapt payload and SSRC to match destination
			for _, audioTrack := range audioTracks[name] {
				packet.Header.PayloadType = audioTrack.PayloadType()
				packet.Header.SSRC = audioTrack.SSRC()
				if writeErr := audioTrack.WriteRTP(packet); writeErr != nil {
					log.Printf("Failed to write to audio track: %s", err)
					continue
				}
			}
		}
	}()

	// Wait for stopped ffmpeg
	if err = ffmpeg.Wait(); err != nil {
		log.Printf("Faited to wait for ffmpeg: %s", err)
	}

	// Close UDP listeners
	if err = videoListener.Close(); err != nil {
		log.Printf("Faited to close UDP listener: %s", err)
	}
	if err = audioListener.Close(); err != nil {
		log.Printf("Faited to close UDP listener: %s", err)
	}
	q.Unregister(videoInput)
}

func startFFmpeg(in <-chan []byte) (ffmpeg *exec.Cmd, err error) {
	ffmpegArgs := []string{"-hide_banner", "-loglevel", "error", "-i", "pipe:0",
		// Audio
		"-vn", "-c:a", "libopus", "-b:a", "160k",
		"-f", "rtp", "rtp://127.0.0.1:5004",
		// Source
		"-an", "-c:v", "copy", "-b:v", "3000k", "-maxrate", "5000k", "-bufsize", "5000k",
		"-f", "rtp", "rtp://127.0.0.1:5005"}
	ffmpeg = exec.Command("ffmpeg", ffmpegArgs...)

	// Handle errors output
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			log.Printf("[WEBRTC FFMPEG %s] %s", "demo", scanner.Text())
		}
	}()

	// Handle stream input
	input, err := ffmpeg.StdinPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		for data := range in {
			if _, err := input.Write(data); err != nil {
				log.Printf("Failed to write data to ffmpeg input: %s", err)
			}
		}

		// End of stream
		ffmpeg.Process.Kill()
	}()

	// Start process
	err = ffmpeg.Start()
	return ffmpeg, err
}
