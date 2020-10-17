package telnet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"

	"gitlab.crans.org/nounous/ghostream/stream"
)

// Convert rawvideo to ANSI text
func streamToTextStream(stream *stream.Stream, text *[]byte, cfg *Options) {
	// Start ffmpeg
	video := make(chan []byte)
	stream.Register(video)
	_, rawvideo, err := startFFmpeg(video, cfg)
	if err != nil {
		log.Printf("Error while starting ffmpeg: %s", err)
	}

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

		*text = textBuff.Bytes()
	}
}

// Start a ffmpeg instance to convert stream into rawvideo
func startFFmpeg(in <-chan []byte, cfg *Options) (*exec.Cmd, *io.ReadCloser, error) {
	bitrate := fmt.Sprintf("%dk", cfg.Width*cfg.Height/cfg.Delay)
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
