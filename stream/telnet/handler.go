package telnet

import (
	"log"
	"net"
	"strings"
	"time"

	"gitlab.crans.org/nounous/ghostream/stream"
)

func handleViewer(s net.Conn, streams map[string]*stream.Stream, cfg *Options) {
	// Prompt user about stream name
	if _, err := s.Write([]byte("[GHOSTREAM]\nEnter stream name: ")); err != nil {
		log.Printf("Error while writing to TCP socket: %s", err)
		s.Close()
		return
	}
	buff := make([]byte, 255)
	n, err := s.Read(buff)
	if err != nil {
		log.Printf("Error while requesting stream ID to telnet client: %s", err)
		s.Close()
		return
	}
	name := strings.TrimSpace(string(buff[:n])) + "@text"
	if len(name) < 1 {
		// Too short, exit
		s.Close()
		return
	}

	// Wait a bit
	time.Sleep(time.Second)

	// Get requested stream
	st, ok := streams[name]
	if !ok {
		log.Println("Stream does not exist, kicking new Telnet viewer")
		if _, err := s.Write([]byte("This stream is inactive.\n")); err != nil {
			log.Printf("Error while writing to TCP socket: %s", err)
		}
		s.Close()
		return
	}

	// Register new client
	log.Printf("New Telnet viewer for stream '%s'", name)
	c := make(chan []byte, 128)
	st.Register(c)
	st.IncrementClientCount()

	// Hide terminal cursor
	if _, err = s.Write([]byte("\033[?25l")); err != nil {
		log.Printf("Error while writing to TCP socket: %s", err)
		s.Close()
		return
	}

	// Receive data and send them
	for data := range c {
		if len(data) < 1 {
			log.Print("Remove Telnet viewer because of end of stream")
			break
		}

		// Send data
		_, err := s.Write(data)
		if err != nil {
			log.Printf("Remove Telnet viewer because of sending error, %s", err)
			break
		}
	}

	// Close output
	st.Unregister(c)
	st.DecrementClientCount()
	s.Close()
}
