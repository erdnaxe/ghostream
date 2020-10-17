package stream

import (
	"testing"
)

func TestWithoutOutputs(t *testing.T) {
	stream := New()
	defer stream.Close()
	stream.Broadcast <- []byte("hello world")
}

func TestWithOneOutput(t *testing.T) {
	stream := New()
	defer stream.Close()

	// Register one output
	output := make(chan []byte, 64)
	stream.Register(output)

	// Try to pass one message
	stream.Broadcast <- []byte("hello world")
	msg := <-output
	if string(msg) != "hello world" {
		t.Errorf("Message has wrong content: %s != hello world", msg)
	}

	// Unregister
	stream.Unregister(output)
}
