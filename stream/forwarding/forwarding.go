// Package forwarding forwards incoming stream to other streaming services
package forwarding

import (
	"bufio"
	"log"
	"os/exec"

	"gitlab.crans.org/nounous/ghostream/messaging"
)

// Options to configure the stream forwarding.
// For each stream name, user can provide several URL to forward stream to
type Options map[string][]string

// Serve handles incoming packets from SRT and forward them to other external services
func Serve(streams *messaging.Streams, cfg Options) {
	if len(cfg) < 1 {
		// No forwarding, ignore
		return
	}

	// Subscribe to new stream event
	event := make(chan string, 8)
	streams.Subscribe(event)
	log.Printf("Stream forwarding initialized")

	// For each new stream
	for name := range event {
		streamCfg, ok := cfg[name]
		if !ok {
			// Not configured
			continue
		}

		// Get stream
		stream, err := streams.Get(name)
		if err != nil {
			log.Printf("Failed to get stream '%s'", name)
		}

		// Get specific quality
		// FIXME: make it possible to forward other qualities
		qualityName := "source"
		quality, err := stream.GetQuality(qualityName)
		if err != nil {
			log.Printf("Failed to get quality '%s'", qualityName)
		}

		// Start forwarding
		log.Printf("Starting forwarding for '%s' quality '%s'", name, qualityName)
		go forward(quality, streamCfg)
	}
}

// Start a FFMPEG instance and redirect stream output to forwarded streams
func forward(q *messaging.Quality, fwdCfg []string) {
	output := make(chan []byte, 1024)
	q.Register(output)

	// Launch FFMPEG instance
	params := []string{"-hide_banner", "-loglevel", "error", "-re", "-i", "pipe:0"}
	for _, url := range fwdCfg {
		params = append(params, "-f", "flv", "-preset", "ultrafast", "-tune", "zerolatency",
			"-c", "copy", url)
	}
	ffmpeg := exec.Command("ffmpeg", params...)

	// Open pipes
	input, err := ffmpeg.StdinPipe()
	if err != nil {
		log.Printf("Error while opening forwarding ffmpeg input pipe: %s", err)
		return
	}
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		log.Printf("Error while opening forwarding ffmpeg output pipe: %s", err)
		return
	}

	// Start FFMpeg
	if err := ffmpeg.Start(); err != nil {
		log.Printf("Error while starting forwarding ffmpeg instance: %s", err)
	}

	// Kill FFMPEG when stream is ended
	defer func() {
		_ = input.Close()
		_ = errOutput.Close()
		_ = ffmpeg.Process.Kill()
		q.Unregister(output)
	}()

	// Log standard error output
	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			log.Printf("[FORWARDING FFMPEG] %s", scanner.Text())
		}
	}()

	// Read stream output and redirect immediately to ffmpeg
	for data := range output {
		_, err := input.Write(data)

		if err != nil {
			log.Printf("Error while writing to forwarded stream: %s", err)
			break
		}
	}
}
