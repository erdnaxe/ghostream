// Package stream defines a structure to communication between inputs and outputs
package stream

import "sync"

// Stream makes packages able to subscribe to an incoming stream
type Stream struct {
	// Incoming data come from this channel
	Broadcast chan<- []byte

	// Use a map to be able to delete an item
	outputs map[chan []byte]struct{}

	// Count clients for statistics
	nbClients int

	// Mutex to lock this ressource
	lock sync.Mutex
}

// New creates a new stream.
func New() *Stream {
	s := &Stream{}
	broadcast := make(chan []byte, 64)
	s.Broadcast = broadcast
	s.outputs = make(map[chan []byte]struct{})
	s.nbClients = 0
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

// Register a new output on a stream.
// If hidden in true, then do not count this client.
func (s *Stream) Register(output chan []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.outputs[output] = struct{}{}
}

// Unregister removes an output.
// If hidden in true, then do not count this client.
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

// ClientCount returns the number of clients
func (s *Stream) ClientCount() int {
	return s.nbClients
}

// IncrementClientCount increments the number of clients
func (s *Stream) IncrementClientCount() {
	s.nbClients++
}

// DecrementClientCount decrements the number of clients
func (s *Stream) DecrementClientCount() {
	s.nbClients--
}
