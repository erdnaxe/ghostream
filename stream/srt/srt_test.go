package srt

import (
	"bufio"
	"os/exec"
	"testing"
	"time"
)

// TestSplitHostPort Try to split a host like 127.0.0.1:1234 in host, port (127.0.0.1, 1234à
func TestSplitHostPort(t *testing.T) {
	// Split 127.0.0.1:1234
	host, port, err := splitHostPort("127.0.0.1:1234")
	if err != nil {
		t.Errorf("Failed to split host and port, %s", err)
	}
	if host != "127.0.0.1" || port != 1234 {
		t.Errorf("splitHostPort returned %v:%d != 127.0.0.1:1234", host, port)
	}

	// Split :1234
	host, port, err = splitHostPort(":1234")
	if err != nil {
		t.Errorf("Failed to split host and port, %s", err)
	}
	if host != "0.0.0.0" || port != 1234 {
		t.Errorf("splitHostPort returned %v:%d != 0.0.0.0:1234", host, port)
	}

	// Split demo, should fail
	host, port, err = splitHostPort("demo")
	if err == nil {
		t.Errorf("splitHostPort managed to split unsplitable hostport")
	}

	// Split demo:port, should fail
	host, port, err = splitHostPort("demo:port")
	if err == nil {
		t.Errorf("splitHostPort managed to split unsplitable hostport")
	}
}

// TestServeSRT Serve a SRT server, stream content during 5 seconds and ensure that it is well received
func TestServeSRT(t *testing.T) {
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

	go Serve(&Options{Enabled: true, ListenAddress: ":9711", MaxClients: 2}, nil, nil, nil)

	ffmpeg := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=size=640x480:rate=10",
		"-f", "flv", "srt://127.0.0.1:9711?streamid=demo:")

	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		t.Fatal("Error while querying ffmpeg output:", err)
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
