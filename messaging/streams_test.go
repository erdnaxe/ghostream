package messaging

import "testing"

func TestWithOneStream(t *testing.T) {
	streams := New()

	// Subscribe to new streams
	event := make(chan string, 8)
	streams.Subscribe(event)

	// Create a stream
	stream, err := streams.Create("demo")
	if err != nil {
		t.Errorf("Failed to create stream")
	}

	// Check that we receive the creation event
	e := <-event
	if e != "demo" {
		t.Errorf("Message has wrong content: %s != demo", e)
	}

	// Create a quality
	quality, err := stream.CreateQuality("source")
	if err != nil {
		t.Errorf("Failed to create quality")
	}

	// Register one output
	output := make(chan []byte, 64)
	quality.Register(output)
	stream.IncrementClientCount()

	// Try to pass one message
	quality.Broadcast <- []byte("hello world")
	msg := <-output
	if string(msg) != "hello world" {
		t.Errorf("Message has wrong content: %s != hello world", msg)
	}

	// Check client count
	if count := stream.ClientCount(); count != 1 {
		t.Errorf("Client counter returned %d, expected 1", count)
	}

	// Unregister
	quality.Unregister(output)
	stream.DecrementClientCount()

	// Check client count
	if count := stream.ClientCount(); count != 0 {
		t.Errorf("Client counter returned %d, expected 0", count)
	}
}
