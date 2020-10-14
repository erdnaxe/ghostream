package telnet

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net"
	"testing"
)

// TestTelnetOutput creates a TCP client that connects to the server and get one image.
func TestTelnetOutput(t *testing.T) {
	// Enable and start Telnet server
	Serve(&Options{
		Enabled:       true,
		ListenAddress: "127.0.0.1:8023",
		Width:         80,
		Height:        45,
		Delay:         50,
	})

	// Generate a random image, that should be given by FFMPEG
	sampleImage := make([]byte, Cfg.Width*Cfg.Height)
	rand.Read(sampleImage)
	reader := ioutil.NopCloser(bytes.NewBuffer(sampleImage))
	// Send the image to the server
	StartASCIIArtStream("demo", reader)

	// Connect to the Telnet server
	client, err := net.Dial("tcp", Cfg.ListenAddress)
	if err != nil {
		t.Fatalf("Error while connecting to the TCP server: %s", err)
	}

	// Create a sufficient large buffer
	buff := make([]byte, 3*len(sampleImage))

	// Read the input, and ensure that it is correct
	// [GHOSTREAM]
	// Enter stream ID:
	n, err := client.Read(buff)
	if err != nil {
		t.Fatalf("Error while reading from TCP: %s", err)
	}
	if n != len("[GHOSTREAM]\nEnter stream ID: ") {
		t.Fatalf("Read %d bytes from TCP, expected %d, read: %s", n, len("[GHOSTREAM]\nEnter stream ID: "), buff[:n])
	}

	// Send stream ID
	_, err = client.Write([]byte("demo"))
	if err != nil {
		t.Fatalf("Error while writing from TCP: %s", err)
	}

	// Read the generated image
	n, err = client.Read(buff)
	if err != nil {
		t.Fatalf("Error while reading the image from TCP: %s", err)
	}
	// Ensure that the size of the image is correct
	if n != 42+(2*Cfg.Width+1)*Cfg.Height {
		t.Fatalf("Read %d from TCP, expected %d", n, 42+(2*Cfg.Width+1)*Cfg.Height)
	}

	if GetNumberConnectedSessions("demo") != 1 {
		t.Fatalf("Expected one telnet client only, found %d", GetNumberConnectedSessions("demo"))
	}
}
