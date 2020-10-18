// Package text transcode a video to text
package text

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"gitlab.crans.org/nounous/ghostream/stream"
)

// Options holds text package configuration
type Options struct {
	Enabled   bool
	Width     int
	Height    int
	Framerate int
}

// Init text transcoder
func Init(streams map[string]*stream.Stream, cfg *Options) {
	if !cfg.Enabled {
		// Text transcode is not enabled, ignore
		return
	}

	// Regulary check existing streams
	for {
		for sourceName, sourceStream := range streams {
			if strings.Contains(sourceName, "@") {
				// Not a source stream, pass
				continue
			}

			// Check that the transcoded stream does not already exist
			name := sourceName + "@text"
			_, ok := streams[name]
			if ok {
				// Stream is already transcoded
				continue
			}

			// Start conversion
			log.Printf("Starting text transcode '%s'", name)
			st := stream.New()
			streams[name] = st

			go transcode(sourceStream, st, cfg)
		}

		// Regulary pull stream list,
		// it may be better to tweak the messaging system
		// to get an event on a new stream.
		time.Sleep(time.Second)
	}
}

// Convert video to ANSI text
func transcode(input, output *stream.Stream, cfg *Options) {
	// Start ffmpeg to transcode video to rawvideo
	videoInput := make(chan []byte, 1024)
	input.Register(videoInput)
	ffmpeg, rawvideo, err := startFFmpeg(videoInput, cfg)
	if err != nil {
		log.Printf("Error while starting ffmpeg: %s", err)
		return
	}

	// Transcode rawvideo to ANSI text
	pixelBuff := make([]byte, cfg.Width*cfg.Height)
	textBuff := bytes.Buffer{}
	for {
		n, err := (*rawvideo).Read(pixelBuff)
		if err != nil {
			log.Printf("An error occurred while reading input: %s", err)
			break
		}
		if n == 0 {
			// Stream is finished
			break
		}

		// Header
		textBuff.Reset()
		textBuff.Grow((40*cfg.Width+6)*cfg.Height + 47)
		for i := 0; i < 42; i++ {
			textBuff.WriteByte('\n')
		}

		// Convert image to ASCII
		for i, pixel := range pixelBuff {
			if i%cfg.Width == 0 {
				// New line
				textBuff.WriteString("\033[49m\n")
			}

			// Print two times the character to make a square
			text := fmt.Sprintf("\033[48;2;%d;%d;%dm ", pixel, pixel, pixel)
			textBuff.WriteString(text)
			textBuff.WriteString(text)
		}
		textBuff.WriteString("\033[49m")

		output.Broadcast <- textBuff.Bytes()
	}

	// Stop transcode
	ffmpeg.Process.Kill()
}

// Start a ffmpeg instance to convert stream into rawvideo
func startFFmpeg(in <-chan []byte, cfg *Options) (*exec.Cmd, *io.ReadCloser, error) {
	bitrate := fmt.Sprintf("%dk", cfg.Width*cfg.Height*cfg.Framerate)
	ffmpegArgs := []string{"-hide_banner", "-loglevel", "error", "-i", "pipe:0",
		"-an", "-vf", fmt.Sprintf("scale=%dx%d", cfg.Width, cfg.Height),
		"-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-bufsize", bitrate,
		"-q", "42", "-pix_fmt", "gray", "-f", "rawvideo", "pipe:1"}
	ffmpeg := exec.Command("ffmpeg", ffmpegArgs...)

	// Handle errors output
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			log.Printf("[TELNET FFMPEG %s] %s", "demo", scanner.Text())
		}
	}()

	// Handle text output
	output, err := ffmpeg.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	// Handle stream input
	input, err := ffmpeg.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	go func() {
		for data := range in {
			input.Write(data)
		}
	}()

	// Start process
	err = ffmpeg.Start()
	return ffmpeg, &output, nil
}
