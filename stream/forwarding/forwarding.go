package forwarding

import (
	"bufio"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
	"io"
	"log"
	"os/exec"
)

// Options to configure the stream forwarding.
// For each stream name, user can provide several URL to forward stream to
type Options map[string][]string

var (
	cfg                Options
	forwardingChannel  chan srt.Packet
	ffmpegInstances    = make(map[string]*exec.Cmd)
	ffmpegInputStreams = make(map[string]*io.WriteCloser)
)

// Serve Load configuration and initialize SRT channel
func Serve(c Options, channel chan srt.Packet) {
	cfg = c
	forwardingChannel = channel
	log.Printf("Stream forwarding initialized")
	waitForPackets()
}

func waitForPackets() {
	for {
		var err error = nil
		packet := <-forwardingChannel
		switch packet.PacketType {
		case "register":
			err = RegisterStream(packet.StreamName)
			break
		case "sendData":
			err = SendPacket(packet.StreamName, packet.Data)
			break
		case "close":
			err = CloseConnection(packet.StreamName)
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

// RegisterStream Declare a new open stream and create ffmpeg instances
func RegisterStream(name string) error {
	streams, exist := cfg[name]
	if !exist || len(streams) == 0 {
		// Nothing to do, not configured
		return nil
	}

	// Launch FFMPEG instance
	params := []string{"-re", "-i", "pipe:0"}
	for _, stream := range streams {
		params = append(params, "-f", "flv", "-preset", "ultrafast", "-tune", "zerolatency",
			"-c", "copy", stream)
	}
	ffmpeg := exec.Command("ffmpeg", params...)

	// Open pipes
	input, err := ffmpeg.StdinPipe()
	if err != nil {
		return err
	}
	output, err := ffmpeg.StdoutPipe()
	if err != nil {
		return err
	}
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		return err
	}
	ffmpegInstances[name] = ffmpeg
	ffmpegInputStreams[name] = &input

	// Start FFMpeg
	if err := ffmpeg.Start(); err != nil {
		return err
	}

	// Log ffmpeg output
	go func() {
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			log.Printf("[FFMPEG %s] %s", name, scanner.Text())
		}
	}()

	// Log also error output
	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			log.Printf("[FFMPEG ERR %s] %s", name, scanner.Text())
		}
	}()

	return nil
}

// SendPacket forward data to all FFMpeg instances related to the stream name
func SendPacket(name string, data []byte) error {
	stdin := ffmpegInputStreams[name]
	if stdin == nil {
		// Don't need to forward stream
		return nil
	}
	_, err := (*stdin).Write(data)
	return err
}

// CloseConnection When the stream is ended, close FFMPEG instances
func CloseConnection(name string) error {
	ffmpeg := ffmpegInstances[name]
	if ffmpeg == nil {
		// No stream to close
		return nil
	}
	if err := ffmpeg.Process.Kill(); err != nil {
		return err
	}
	delete(ffmpegInstances, name)
	delete(ffmpegInputStreams, name)
	return nil
}
