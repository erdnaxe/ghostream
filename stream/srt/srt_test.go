package srt

import (
	"bufio"
	"log"
	"os/exec"
	"testing"
	"time"
)

// TestSplitHostPort Try to split a host like 127.0.0.1:1234 in host, port (127.0.0.1, 1234à
func TestSplitHostPort(t *testing.T) {
	host, port := splitHostPort("127.0.0.1:1234")
	if host != "127.0.0.1" || port != 1234 {
		t.Errorf("splitHostPort returned %v:%d != 127.0.0.1:1234", host, port)
	}

	host, port = splitHostPort(":1234")
	if host != "0.0.0.0" || port != 1234 {
		t.Errorf("splitHostPort returned %v:%d != 0.0.0.0:1234", host, port)
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

	go Serve(&Options{ListenAddress: ":9711", MaxClients: 2}, nil, nil, nil)

	ffmpeg := exec.Command("ffmpeg",
		"-i", "http://ftp.crans.org/events/Blender%20OpenMovies/big_buck_bunny_480p_stereo.ogg",
		"-f", "flv", "srt://127.0.0.1:9711?streamid=demo:")

	output, err := ffmpeg.StdoutPipe()
	errOutput, err := ffmpeg.StderrPipe()
	if err != nil {
		t.Fatal("Error while querying ffmpeg output:", err)
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

	// TODO Check that the stream ran
	// TODO Kill SRT server
}
