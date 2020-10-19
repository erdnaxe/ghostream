// Package messaging defines a structure to communication between inputs and outputs
package messaging

import (
	"errors"
	"log"
	"sync"
)

// Streams hold all application streams.
type Streams struct {
	// Associate each stream name to the stream
	streams map[string]*Stream

	// Mutex to lock streams
	lockStreams sync.Mutex

	// Subscribers get notified when a new stream is created
	// Use a map to be able to delete a subscriber
	eventSubscribers map[chan string]struct{}

	// Mutex to lock eventSubscribers
	lockSubscribers sync.Mutex
}

// New creates a new stream list.
func New() (l *Streams) {
	l = &Streams{}
	l.streams = make(map[string]*Stream)
	l.eventSubscribers = make(map[chan string]struct{})
	return l
}

// Subscribe to get notified on new stream.
func (l *Streams) Subscribe(output chan string) {
	l.lockSubscribers.Lock()
	l.eventSubscribers[output] = struct{}{}
	l.lockSubscribers.Unlock()
}

// Unsubscribe to no longer get notified on new stream.
func (l *Streams) Unsubscribe(output chan string) {
	// Make sure we did not already delete this subscriber
	l.lockSubscribers.Lock()
	if _, ok := l.eventSubscribers[output]; ok {
		delete(l.eventSubscribers, output)
	}
	l.lockSubscribers.Unlock()
}

// Create a new stream.
func (l *Streams) Create(name string) (s *Stream, err error) {
	// If stream already exist, fail
	if _, ok := l.streams[name]; ok {
		return nil, errors.New("stream already exists")
	}

	// Create stream
	s = newStream()
	l.lockStreams.Lock()
	l.streams[name] = s
	l.lockStreams.Unlock()

	// Notify
	l.lockSubscribers.Lock()
	for sub := range l.eventSubscribers {
		select {
		case sub <- name:
		default:
			log.Printf("Failed to announce stream '%s' to subscriber", name)
		}
	}
	l.lockSubscribers.Unlock()
	return s, nil
}

// Get a stream.
func (l *Streams) Get(name string) (s *Stream, err error) {
	// If stream does exist, return it
	l.lockStreams.Lock()
	s, ok := l.streams[name]
	l.lockStreams.Unlock()
	if !ok {
		return nil, errors.New("stream does not exist")
	}
	return s, nil
}

// Delete a stream.
func (l *Streams) Delete(name string) {
	// Make sure we did not already delete this stream
	l.lockStreams.Lock()
	if _, ok := l.streams[name]; ok {
		delete(l.streams, name)
	}
	l.lockStreams.Unlock()
}
