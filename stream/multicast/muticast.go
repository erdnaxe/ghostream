package multicast

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
)

var (
	ffmpegInstances    = make(map[string][]*exec.Cmd)
	ffmpegInputStreams = make(map[string][]io.WriteCloser)
)

// Declare a new open stream and create ffmpeg instances
func RegisterStream(streamKey string) {
	ffmpegInstances[streamKey] = []*exec.Cmd{}
	ffmpegInputStreams[streamKey] = []io.WriteCloser{}

	// TODO Export the list of multicasts
	for _, stream := range []string{fmt.Sprintf("rtmp://live.twitch.tv/app/%s", "TWITCH_STREAM_KEY")} {
		// Launch FFMPEG instance
		// TODO Set optimal parameters
		ffmpeg := exec.Command("ffmpeg", "-re", "-i", "pipe:0", "-f", "flv", "-c:v", "libx264", "-preset",
			"veryfast", "-maxrate", "3000k", "-bufsize", "6000k", "-pix_fmt", "yuv420p", "-g", "50", "-c:a", "aac",
			"-b:a", "160k", "-ac", "2", "-ar", "44100", stream)
		ffmpegInstances[streamKey] = append(ffmpegInstances[streamKey], ffmpeg)
		input, _ := ffmpeg.StdinPipe()
		ffmpegInputStreams[streamKey] = append(ffmpegInputStreams[streamKey], input)
		output, _ := ffmpeg.StdoutPipe()

		if err := ffmpeg.Start(); err != nil {
			panic(err)
		}

		// Log ffmpeg output
		go func() {
			scanner := bufio.NewScanner(output)
			for scanner.Scan() {
				log.Println(scanner.Text())
			}
		}()
	}
}

// When a SRT packet is received, transmit it to all FFMPEG instances related to the stream key
func SendPacket(streamKey string, data []byte) {
	for _, stdin := range ffmpegInputStreams[streamKey] {
		_, err := stdin.Write(data)
		if err != nil {
			panic(err)
		}
	}

}

// When the stream is ended, close FFMPEG instances
func CloseConnection(streamKey string) {
	for _, ffmpeg := range ffmpegInstances[streamKey] {
		if err := ffmpeg.Process.Kill(); err != nil {
			panic(err)
		}
	}
	delete(ffmpegInstances, streamKey)
	delete(ffmpegInputStreams, streamKey)
}
