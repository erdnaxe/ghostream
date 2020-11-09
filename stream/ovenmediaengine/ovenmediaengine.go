// Package ovenmediaengine provides the forwarding to an ovenmediaengine server to handle the web client
package ovenmediaengine

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"gitlab.crans.org/nounous/ghostream/messaging"
)

// Options holds ovenmediaengine package configuration
type Options struct {
	Enabled bool
	URL     string
	App     string
}

var (
	cfg *Options
)

// Serve handles incoming packets from SRT and forward them to OME
func Serve(streams *messaging.Streams, cfg *Options) {
	if !cfg.Enabled {
		return
	}

	// Subscribe to new stream event
	event := make(chan string, 8)
	streams.Subscribe(event)
	log.Printf("Stream forwarding to OME initialized")

	// For each new stream
	for name := range event {
		// Get stream
		stream, err := streams.Get(name)
		if err != nil {
			log.Printf("Failed to get stream '%s'", name)
			return
		}

		qualityName := "source"
		quality, err := stream.GetQuality(qualityName)
		if err != nil {
			log.Printf("Failed to get quality '%s'", qualityName)
		}

		// Start forwarding
		log.Printf("Starting forwarding to OME for '%s'", name)
		go forward(name, quality)
	}
}

// Start a FFMPEG instance and redirect stream output to OME
func forward(name string, q *messaging.Quality) {
	output := make(chan []byte, 1024)
	q.Register(output)

	// TODO When a new OME version got released with SRT support, directly forward SRT packets, without using unwanted RTMP transport
	// Launch FFMPEG instance
	params := []string{"-hide_banner", "-loglevel", "error", "-i", "pipe:0", "-f", "flv", "-c:v", "copy",
		"-c:a", "aac", "-b:a", "160k", "-ar", "44100",
		fmt.Sprintf("rtmp://%s/%s/%s", cfg.URL, cfg.App, name)}
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
			log.Printf("[FORWARDING OME FFMPEG %s] %s", name, scanner.Text())
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
