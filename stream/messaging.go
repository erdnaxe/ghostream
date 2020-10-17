// Package stream defines a structure to communication between inputs and outputs
package stream

import "sync"

// Stream makes packages able to subscribe to an incoming stream
type Stream struct {
	// Incoming data come from this channel
	Broadcast chan<- []byte

	// Use a map to be able to delete an item
	outputs map[chan []byte]struct{}

	// Mutex to lock this ressource
	lock sync.Mutex
}

// New creates a new stream.
func New() *Stream {
	s := &Stream{}
	broadcast := make(chan []byte, 64)
	s.Broadcast = broadcast
	s.outputs = make(map[chan []byte]struct{})
	go s.run(broadcast)
	return s
}

func (s *Stream) run(broadcast <-chan []byte) {
	for msg := range broadcast {
		func() {
			s.lock.Lock()
			defer s.lock.Unlock()
			for output := range s.outputs {
				select {
				case output <- msg:
				default:
					// If full, do a ring buffer
					<-output
					output <- msg
				}
			}
		}()
	}

	// Incoming chan has been closed, close all outputs
	s.lock.Lock()
	defer s.lock.Unlock()
	for ch := range s.outputs {
		delete(s.outputs, ch)
		close(ch)
	}
}

// Close the incoming chan, this will also delete all outputs
func (s *Stream) Close() {
	close(s.Broadcast)
}

// Register a new output on a stream
func (s *Stream) Register(output chan []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.outputs[output] = struct{}{}
}

// Unregister removes an output
func (s *Stream) Unregister(output chan []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Make sure we did not already close this output
	_, ok := s.outputs[output]
	if ok {
		delete(s.outputs, output)
		close(output)
	}
}

// Count number of outputs
func (s *Stream) Count() int {
	return len(s.outputs)
}
