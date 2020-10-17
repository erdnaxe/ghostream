package telnet

import (
	"log"
	"net"
	"strings"
	"time"

	"gitlab.crans.org/nounous/ghostream/stream"
)

func handleViewer(s net.Conn, streams map[string]*stream.Stream, textStreams map[string]*[]byte, cfg *Options) {
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
	name := strings.TrimSpace(string(buff[:n]))
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
	log.Printf("New Telnet viewer for stream %s", name)
	st.IncrementClientCount()

	// Hide terminal cursor
	if _, err = s.Write([]byte("\033[?25l")); err != nil {
		log.Printf("Error while writing to TCP socket: %s", err)
		s.Close()
		return
	}

	// Send stream
	for {
		text, ok := textStreams[name]
		if !ok {
			log.Println("Stream is not converted to text, kicking Telnet viewer")
			if _, err := s.Write([]byte("This stream cannot be opened.\n")); err != nil {
				log.Printf("Error while writing to TCP socket: %s", err)
			}
			break
		}

		// Send text to client
		n, err := s.Write(*text)
		if err != nil || n == 0 {
			log.Printf("Error while sending TCP data: %s", err)
			break
		}

		time.Sleep(time.Duration(cfg.Delay) * time.Millisecond)
	}

	// Close connection
	s.Close()
	st.DecrementClientCount()
}
