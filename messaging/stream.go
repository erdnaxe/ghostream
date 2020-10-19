// Package messaging defines a structure to communication between inputs and outputs
package messaging

import (
	"sync"
)

// Stream makes packages able to subscribe to an incoming stream
type Stream struct {
	// Different qualities of this stream
	qualities map[string]*Quality

	// Mutex to lock outputs map
	lockQualities sync.Mutex

	// Count clients for statistics
	nbClients int
}

func newStream() (s *Stream) {
	s = &Stream{}
	s.qualities = make(map[string]*Quality)
	s.nbClients = 0
	return s
}

// CreateQuality creates a new quality associated with this stream.
func (s *Stream) CreateQuality(name string) (quality *Quality) {
	s.lockQualities.Lock()
	quality = newQuality()
	s.qualities[name] = quality
	s.lockQualities.Unlock()
	return quality
}

// DeleteQuality removes a quality.
func (s *Stream) DeleteQuality(name string) {
	// Make sure we did not already close this output
	s.lockQualities.Lock()
	if _, ok := s.qualities[name]; ok {
		s.qualities[name].Close()
		delete(s.qualities, name)
	}
	s.lockQualities.Unlock()
}

// ClientCount returns the number of clients.
func (s *Stream) ClientCount() int {
	return s.nbClients
}

// IncrementClientCount increments the number of clients.
func (s *Stream) IncrementClientCount() {
	s.nbClients++
}

// DecrementClientCount decrements the number of clients.
func (s *Stream) DecrementClientCount() {
	s.nbClients--
}
