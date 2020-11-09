// Package forwarding forwards incoming stream to other streaming services
package forwarding

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

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
			return
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
		go forward(name, quality, streamCfg)
	}
}

// Start a FFMPEG instance and redirect stream output to forwarded streams
func forward(streamName string, q *messaging.Quality, fwdCfg []string) {
	output := make(chan []byte, 1024)
	q.Register(output)

	// Launch FFMPEG instance
	params := []string{"-hide_banner", "-loglevel", "error", "-i", "pipe:0"}
	for _, url := range fwdCfg {
		// If the url should be date-formatted, replace special characters with the current time information
		now := time.Now()
		formattedURL := strings.ReplaceAll(url, "%Y", fmt.Sprintf("%04d", now.Year()))
		formattedURL = strings.ReplaceAll(formattedURL, "%m", fmt.Sprintf("%02d", now.Month()))
		formattedURL = strings.ReplaceAll(formattedURL, "%d", fmt.Sprintf("%02d", now.Day()))
		formattedURL = strings.ReplaceAll(formattedURL, "%H", fmt.Sprintf("%02d", now.Hour()))
		formattedURL = strings.ReplaceAll(formattedURL, "%M", fmt.Sprintf("%02d", now.Minute()))
		formattedURL = strings.ReplaceAll(formattedURL, "%S", fmt.Sprintf("%02d", now.Second()))
		formattedURL = strings.ReplaceAll(formattedURL, "%name", streamName)

		params = append(params, "-f", "flv",
			"-c:v", "copy", "-c:a", "aac", "-b:a", "160k", "-ar", "44100", formattedURL)
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
			log.Printf("[FORWARDING FFMPEG %s] %s", streamName, scanner.Text())
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
