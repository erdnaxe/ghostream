// Package srt serves a SRT server
package srt

import (
	"log"

	"github.com/haivision/srtgo"
	"gitlab.crans.org/nounous/ghostream/stream"
)

func handleStreamer(socket *srtgo.SrtSocket, streams map[string]*stream.Stream, name string) {
	// Check stream does not exist
	if _, ok := streams[name]; ok {
		log.Print("Stream already exists, refusing new streamer")
		socket.Close()
		return
	}

	// Create stream
	log.Printf("New SRT streamer for stream %s", name)
	st := stream.New()
	streams[name] = st

	// Read RTP packets forever and send them to the WebRTC Client
	for {
		// Create a new buffer
		// UDP packet cannot be larger than MTU (1500)
		buff := make([]byte, 1500)

		// 5s timeout
		n, err := socket.Read(buff, 5000)
		if err != nil {
			log.Println("Error occurred while reading SRT socket:", err)
			break
		}

		if n == 0 {
			// End of stream
			log.Printf("Received no bytes, stopping stream.")
			break
		}

		// Send raw data to other streams
		buff = buff[:n]
		st.Broadcast <- buff
	}

	// Close stream
	st.Close()
	socket.Close()
	delete(streams, name)
}

func handleViewer(s *srtgo.SrtSocket, streams map[string]*stream.Stream, name string) {
	log.Printf("New SRT viewer for stream %s", name)

	// Get requested stream
	st, ok := streams[name]
	if !ok {
		log.Println("Stream does not exist, refusing new viewer")
		return
	}

	// Register new output
	c := make(chan []byte, 1024)
	st.Register(c)
	st.IncrementClientCount()

	// Receive data and send them
	for data := range c {
		if len(data) < 1 {
			log.Print("Remove SRT viewer because of end of stream")
			break
		}

		// Send data
		_, err := s.Write(data, 1000)
		if err != nil {
			log.Printf("Remove SRT viewer because of sending error, %s", err)
			break
		}
	}

	// Close output
	st.Unregister(c)
	st.DecrementClientCount()
	s.Close()
}
