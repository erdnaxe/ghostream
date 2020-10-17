package stream

import (
	"testing"
)

func TestWithoutOutputs(t *testing.T) {
	stream := New()
	defer stream.Close()
	stream.Broadcast <- "hello world"
}

func TestWithOneOutput(t *testing.T) {
	stream := New()
	defer stream.Close()

	// Register one output
	output := make(chan interface{}, 64)
	stream.Register(output)

	// Try to pass one message
	stream.Broadcast <- "hello world"
	msg := <-output
	if m, ok := msg.(string); !ok || m != "hello world" {
		t.Error("Message has wrong type or content")
	}

	// Unregister
	stream.Unregister(output)
}
