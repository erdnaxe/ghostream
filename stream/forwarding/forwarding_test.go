package forwarding

import (
	"bufio"
	"os/exec"
	"testing"
	"time"

	"gitlab.crans.org/nounous/ghostream/stream/srt"
)

// TestServeSRT Serve a SRT server, stream content during 5 seconds and ensure that it is well received
func TestForwardStream(t *testing.T) {
	// Check that ffmpeg is installed
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

	// Start virtual RTMP server with ffmpeg
	forwardedFfmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-y", // allow overwrite /dev/null
		"-listen", "1", "-i", "rtmp://127.0.0.1:1936/live/app", "-f", "null", "-c", "copy", "/dev/null")
	forwardingErrOutput, err := forwardedFfmpeg.StderrPipe()
	if err != nil {
		t.Fatal("Error while querying ffmpeg forwardingOutput:", err)
	}
	if err := forwardedFfmpeg.Start(); err != nil {
		t.Fatal("Error while starting forwarding stream ffmpeg instance:", err)
	}

	go func() {
		scanner := bufio.NewScanner(forwardingErrOutput)
		for scanner.Scan() {
			t.Fatalf("ffmpeg virtual RTMP server returned %s", scanner.Text())
		}
	}()

	forwardingList := make(map[string][]string)
	forwardingList["demo"] = []string{"rtmp://127.0.0.1:1936/live/app"}

	forwardingChannel := make(chan srt.Packet)

	// Register forwarding stream list
	go Serve(forwardingChannel, forwardingList)

	// Serve SRT Server without authentification backend
	go srt.Serve(&srt.Options{ListenAddress: ":9712", MaxClients: 2}, nil, forwardingChannel, nil)

	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-re", "-f", "lavfi", "-i", "testsrc=size=640x480:rate=10",
		"-f", "flv", "srt://127.0.0.1:9712?streamid=demo:")

	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		t.Fatal("Error while querying ffmpeg forwardingOutput:", err)
	}

	if err := ffmpeg.Start(); err != nil {
		t.Fatal("Error while starting ffmpeg:", err)
	}

	go func() {
		scanner := bufio.NewScanner(errOutput)
		for scanner.Scan() {
			t.Fatalf("ffmpeg virtual source returned %s", scanner.Text())
		}
	}()

	time.Sleep(5 * time.Second) // Delay is in nanoseconds, here 5s

	// TODO Kill SRT server
}
