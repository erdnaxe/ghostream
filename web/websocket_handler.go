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
	webRtcSdp webrtc.SessionDescription
	stream    string
	quality   string
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
			return
		}

		// Get requested stream
		stream, err := streams.Get(c.stream)
		if err != nil {
			log.Printf("Stream not found: %s", c.stream)
			return
		}

		// Get requested quality
		q, err := stream.GetQuality(c.quality)
		if err != nil {
			log.Printf("Quality not found: %s", c.quality)
			return
		}

		// Exchange session descriptions with WebRTC stream server
		// FIXME: Add trickle ICE support
		q.WebRtcRemoteSdp <- c.webRtcSdp
		localDescription := <-q.WebRtcLocalSdp

		// Send new local description
		if err := conn.WriteJSON(localDescription); err != nil {
			log.Println(err)
			return
		}
	}
}
