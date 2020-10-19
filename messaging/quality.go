// Package messaging defines a structure to communication between inputs and outputs
package messaging

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

// Quality holds a specific stream quality.
// It makes packages able to subscribe to an incoming stream.
type Quality struct {
	// Incoming data come from this channel
	Broadcast chan<- []byte

	// Incoming data will be outputed to all those outputs.
	// Use a map to be able to delete an item.
	outputs map[chan []byte]struct{}

	// Mutex to lock outputs map
	lockOutputs sync.Mutex

	// WebRTC session descriptor exchange.
	// When new client connects, a SDP arrives on WebRtcRemoteSdp,
	// then webrtc package answers on WebRtcLocalSdp.
	WebRtcLocalSdp  chan webrtc.SessionDescription
	WebRtcRemoteSdp chan webrtc.SessionDescription
}

func newQuality() (q *Quality) {
	q = &Quality{}
	broadcast := make(chan []byte, 1024)
	q.Broadcast = broadcast
	q.outputs = make(map[chan []byte]struct{})
	q.WebRtcLocalSdp = make(chan webrtc.SessionDescription, 1)
	q.WebRtcRemoteSdp = make(chan webrtc.SessionDescription, 1)
	go q.run(broadcast)
	return q
}

func (q *Quality) run(broadcast <-chan []byte) {
	for msg := range broadcast {
		q.lockOutputs.Lock()
		for output := range q.outputs {
			select {
			case output <- msg:
			default:
				// If full, do a ring buffer
				// Check that output is not of size zero
				if len(output) > 1 {
					<-output
				}
			}
		}
		q.lockOutputs.Unlock()
	}

	// Incoming chan has been closed, close all outputs
	q.lockOutputs.Lock()
	for ch := range q.outputs {
		delete(q.outputs, ch)
		close(ch)
	}
	q.lockOutputs.Unlock()
}

// Close the incoming chan, this will also delete all outputs.
func (q *Quality) Close() {
	close(q.Broadcast)
}

// Register a new output on a stream.
func (q *Quality) Register(output chan []byte) {
	q.lockOutputs.Lock()
	q.outputs[output] = struct{}{}
	q.lockOutputs.Unlock()
}

// Unregister removes an output.
func (q *Quality) Unregister(output chan []byte) {
	// Make sure we did not already close this output
	q.lockOutputs.Lock()
	_, ok := q.outputs[output]
	if ok {
		delete(q.outputs, output)
		close(output)
	}
	defer q.lockOutputs.Unlock()
}
