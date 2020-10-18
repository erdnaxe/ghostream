// Package forwarding forwards incoming stream to other streaming services
package forwarding

import (
	"bufio"
	"log"
	"os/exec"
	"time"

	"gitlab.crans.org/nounous/ghostream/stream"
)

// Options to configure the stream forwarding.
// For each stream name, user can provide several URL to forward stream to
type Options map[string][]string

// Serve handles incoming packets from SRT and forward them to other external services
func Serve(streams map[string]*stream.Stream, cfg Options) {
	if len(cfg) < 1 {
		// No forwarding, ignore
		return
	}

	log.Printf("Stream forwarding initialized")
	for {
		for name, st := range streams {
			fwdCfg, ok := cfg[name]
			if !ok {
				// Not configured
				continue
			}

			// Start forwarding
			log.Printf("Starting forwarding for '%s'", name)
			go forward(st, fwdCfg)
		}

		// Regulary pull stream list,
		// it may be better to tweak the messaging system
		// to get an event on a new stream.
		time.Sleep(time.Second)
	}
}

// Start a FFMPEG instance and redirect stream output to forwarded streams
func forward(st *stream.Stream, fwdCfg []string) {
	output := make(chan []byte, 1024)
	st.Register(output)

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
		st.Unregister(output)
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
