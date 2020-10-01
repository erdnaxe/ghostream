package forwarding

import (
	"bufio"
	"gitlab.crans.org/nounous/ghostream/stream/srt"
	"log"
	"os/exec"
	"testing"
	"time"
)

// TestServeSRT Serve a SRT server, stream content during 5 seconds and ensure that it is well received
func TestForwardStream(t *testing.T) {
	which := exec.Command("which", "ffmpeg")
	if err := which.Start(); err != nil {
		t.Fatal("Error while checking if ffmpeg got installed:", err)
	}
	state, err := which.Process.Wait()
	if err != nil {
		t.Fatal("Error while checking if ffmpeg got installed:", err)
	}
	if state.ExitCode() != 0 {
		// FFMPEG is not installed
		t.Skip("WARNING: FFMPEG is not installed. Skipping stream test")
	}

	forwardedFfmpeg := exec.Command("ffmpeg",
		"-f", "flv", "-listen", "1", "-i", "rtmp://127.0.0.1:1936/live/app", "-c", "copy", "/dev/null")
	forwardingOutput, err := forwardedFfmpeg.StdoutPipe()
	forwardingErrOutput, err := forwardedFfmpeg.StderrPipe()
	if err != nil {
		t.Fatal("Error while querying ffmpeg forwardingOutput:", err)
	}

	if err := forwardedFfmpeg.Start(); err != nil {
		t.Fatal("Error while starting forwarding stream ffmpeg instance:", err)
	}

	go func() {
		scanner := bufio.NewScanner(forwardingOutput)
		for scanner.Scan() {
			log.Printf("[FFMPEG FORWARD TEST] %s", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(forwardingErrOutput)
		for scanner.Scan() {
			log.Printf("[FFMPEG FORWARD ERR TEST] %s", scanner.Text())
		}
	}()

	forwardingList := make(map[string][]string)
	forwardingList["demo"] = []string{"rtmp://127.0.0.1:1936/live/app"}

	forwardingChannel = make(chan srt.Packet)

	// Register forwarding stream list
	Serve(forwardingList, forwardingChannel)

	// Serve HTTP Server
	go srt.Serve(&srt.Options{ListenAddress: ":9712", MaxClients: 2}, forwardingChannel)

	ffmpeg := exec.Command("ffmpeg",
		"-i", "http://ftp.crans.org/events/Blender%20OpenMovies/big_buck_bunny_480p_stereo.ogg",
		"-f", "flv", "srt://127.0.0.1:9712")

	output, err := ffmpeg.StdoutPipe()
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		t.Fatal("Error while querying ffmpeg forwardingOutput:", err)
	}

	if err := ffmpeg.Start(); err != nil {
		t.Fatal("Error while starting ffmpeg:", err)
	}

	go func() {
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			log.Printf("[FFMPEG TEST] %s", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			log.Printf("[FFMPEG ERR TEST] %s", scanner.Text())
		}
	}()

	time.Sleep(5000000000) // Delay is in nanoseconds, here 5s

	if ffmpegInputStreams["demo"] == nil {
		t.Errorf("Stream forwarding does not appear to be working")
	}

	// TODO Check that the stream ran
	// TODO Kill SRT server
}
