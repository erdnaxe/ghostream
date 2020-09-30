package forwarding

import (
	"bufio"
	"io"
	"log"
	"os/exec"
)

// Options to configure the stream forwarding.
// For each stream name, user can provide several URL to forward stream to
type Options map[string][]string

var (
	options            Options
	ffmpegInstances    = make(map[string]*exec.Cmd)
	ffmpegInputStreams = make(map[string]*io.WriteCloser)
)

// New Load configuration
func New(cfg *Options) error {
	options = *cfg
	return nil
}

// RegisterStream Declare a new open stream and create ffmpeg instances
func RegisterStream(name string) {
	if len(options[name]) == 0 {
		return
	}

	params := []string{"-re", "-i", "pipe:0"}
	for _, stream := range options[name] {
		params = append(params, "-f", "flv", "-preset", "ultrafast", "-tune", "zerolatency",
			"-c", "copy", stream)
	}
	// Launch FFMPEG instance
	ffmpeg := exec.Command("ffmpeg", params...)

	// Open pipes
	input, err := ffmpeg.StdinPipe()
	if err != nil {
		panic(err)
	}
	output, err := ffmpeg.StdoutPipe()
	if err != nil {
		panic(err)
	}
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		panic(err)
	}
	ffmpegInstances[name] = ffmpeg
	ffmpegInputStreams[name] = &input

	if err := ffmpeg.Start(); err != nil {
		panic(err)
	}

	// Log ffmpeg output
	go func() {
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			log.Println("[FFMPEG " + name + "] " + scanner.Text())
		}
	}()

	// Log also error output
	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			log.Println("[FFMPEG ERROR " + name + "] " + scanner.Text())
		}
	}()
}

// SendPacket When a SRT packet is received, transmit it to all FFMPEG instances related to the stream key
func SendPacket(name string, data []byte) {
	stdin := ffmpegInputStreams[name]
	_, err := (*stdin).Write(data)
	if err != nil {
		log.Printf("Error while sending a packet to external streaming server for key %s: %s", name, err)
	}
}

// CloseConnection When the stream is ended, close FFMPEG instances
func CloseConnection(name string) {
	ffmpeg := ffmpegInstances[name]
	if err := ffmpeg.Process.Kill(); err != nil {
		panic(err)
	}
	delete(ffmpegInstances, name)
	delete(ffmpegInputStreams, name)
}
