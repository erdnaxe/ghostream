package telnet

import (
	"testing"

	"gitlab.crans.org/nounous/ghostream/stream"
)

// TestTelnetOutput creates a TCP client that connects to the server and get one image.
func TestTelnetOutput(t *testing.T) {
	// Try to start Telnet server while it is disabled
	streams := make(map[string]*stream.Stream)
	go Serve(streams, &Options{Enabled: false})

	// FIXME test connect

	// Enable and start Telnet server
	cfg := Options{
		Enabled:       true,
		ListenAddress: "127.0.0.1:8023",
	}
	go Serve(streams, &cfg)

	// FIXME test connect

	// Generate a random image, that should be given by FFMPEG
	/*sampleImage := make([]byte, cfg.Width*cfg.Height)
	rand.Read(sampleImage)
	reader := ioutil.NopCloser(bytes.NewBuffer(sampleImage))

	// Connect to the Telnet server
	client, err := net.Dial("tcp", cfg.ListenAddress)
	if err != nil {
		t.Fatalf("Error while connecting to the TCP server: %s", err)
	}

	// Say goodbye
	_, err = client.Write([]byte("exit"))
	if err != nil {
		t.Fatalf("Error while closing TCP connection: %s", err)
	}

	client, err = net.Dial("tcp", cfg.ListenAddress)
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

	// Send wrong stream ID
	_, err = client.Write([]byte("toto"))
	if err != nil {
		t.Fatalf("Error while writing from TCP: %s", err)
	}
	n, err = client.Read(buff)
	if err != nil {
		t.Fatalf("Error while reading from TCP: %s", err)
	}
	if n == len("Unknown stream ID.\n") {
		_, err = client.Read(buff)
		if err != nil {
			t.Fatalf("Error while reading from TCP: %s", err)
		}
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
	if n < 10800 {
		t.Fatalf("Read %d from TCP, expected more than 10800", n)
	}

	if GetNumberConnectedSessions("demo") != 1 {
		t.Fatalf("Expected one telnet client only, found %d", GetNumberConnectedSessions("demo"))
	}

	// Close connection, ensure that the counter got decremented
	err = client.Close()
	if err != nil {
		t.Fatalf("Error while closing telnet connection: %s", err)
	}
	// Wait for timeout
	time.Sleep(time.Second)
	if GetNumberConnectedSessions("demo") != 0 {
		t.Fatalf("Expected no telnet client, found %d", GetNumberConnectedSessions("demo"))
	}*/
}
