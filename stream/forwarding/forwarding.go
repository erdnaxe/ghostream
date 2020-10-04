package forwarding

import (
	"bufio"
	"io"
	"log"
	"os/exec"

	"gitlab.crans.org/nounous/ghostream/stream/srt"
)

// Options to configure the stream forwarding.
// For each stream name, user can provide several URL to forward stream to
type Options map[string][]string

// Serve handles incoming packets from SRT and forward them to other external services
func Serve(inputChannel chan srt.Packet, cfg Options) {
	log.Printf("Stream forwarding initialized")
	ffmpegInstances := make(map[string]*exec.Cmd)
	ffmpegInputStreams := make(map[string]*io.WriteCloser)
	for {
		var err error = nil
		// Wait for packets
		packet := <-inputChannel
		switch packet.PacketType {
		case "register":
			err = registerStream(packet.StreamName, ffmpegInstances, ffmpegInputStreams, cfg)
			break
		case "sendData":
			err = sendPacket(packet.StreamName, ffmpegInputStreams, packet.Data)
			break
		case "close":
			err = close(packet.StreamName, ffmpegInstances, ffmpegInputStreams)
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

// registerStream creates ffmpeg instance associated with newly created stream
func registerStream(name string, ffmpegInstances map[string]*exec.Cmd, ffmpegInputStreams map[string]*io.WriteCloser, cfg Options) error {
	streams, exist := cfg[name]
	if !exist || len(streams) == 0 {
		// Nothing to do, not configured
		return nil
	}

	// Launch FFMPEG instance
	params := []string{"-hide_banner", "-loglevel", "error", "-re", "-i", "pipe:0"}
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

// sendPacket forwards data to the ffmpeg instance related to the stream name
func sendPacket(name string, ffmpegInputStreams map[string]*io.WriteCloser, data []byte) error {
	stdin := ffmpegInputStreams[name]
	if stdin == nil {
		// Don't need to forward stream
		return nil
	}
	_, err := (*stdin).Write(data)
	return err
}

// close ffmpeg instance associated with stream name
func close(name string, ffmpegInstances map[string]*exec.Cmd, ffmpegInputStreams map[string]*io.WriteCloser) error {
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
