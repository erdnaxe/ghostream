// Package web serves the JavaScript player and WebRTC negotiation
package web

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"gitlab.crans.org/nounous/ghostream/stream/webrtc"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// clientDescription is sent by new client
type clientDescription struct {
	WebRtcSdp webrtc.SessionDescription
	Stream    string
	Quality   string
}

// websocketHandler exchanges WebRTC SDP and viewer count
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade client connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade client to websocket: %s", err)
		return
	}

	for {
		// Get client description
		c := &clientDescription{}
		err = conn.ReadJSON(c)
		if err != nil {
			log.Printf("Failed to receive client description: %s", err)
			continue
		}

		// Get requested stream
		stream, err := streams.Get(c.Stream)
		if err != nil {
			log.Printf("Stream not found: %s", c.Stream)
			continue
		}

		// Get requested quality
		q, err := stream.GetQuality(c.Quality)
		if err != nil {
			log.Printf("Quality not found: %s", c.Quality)
			continue
		}

		// Exchange session descriptions with WebRTC stream server
		// FIXME: Add trickle ICE support
		q.WebRtcRemoteSdp <- c.WebRtcSdp
		localDescription := <-q.WebRtcLocalSdp

		// Send new local description
		if err := conn.WriteJSON(localDescription); err != nil {
			log.Println(err)
			continue
		}
	}
}
