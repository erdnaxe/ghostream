// Package telnet provides some fancy tools, like an ASCII-art stream.
package telnet

import (
	"io"
	"log"
	"net"
	"time"
)

func asciiChar(pixel byte) string {
	asciiChars := []string{"@", "#", "$", "%", "?", "*", "+", ";", ":", ",", "."}
	return asciiChars[pixel/25]
}

// ServeAsciiArt starts a telnet server that send all packets as ASCII Art
func ServeAsciiArt(reader io.Reader) {
	listener, err := net.Listen("tcp", ":4242")
	if err != nil {
		log.Printf("Error while listening to the port 4242: %s", err)
		return
	}

	currentMessage := ""

	go func() {
		for {
			s, err := listener.Accept()
			if err != nil {
				log.Printf("Error while accepting TCP socket: %s", s)
				continue
			}
			go func(s net.Conn) {
				for {
					n, err := s.Write([]byte(currentMessage))
					if err != nil {
						log.Printf("Error while sending TCP data: %s", err)
						_ = s.Close()
						break
					}
					if n == 0 {
						_ = s.Close()
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			}(s)
		}
	}()

	buff := make([]byte, 2048)
	header := "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"
	for {
		n, _ := reader.Read(buff)
		if n == 0 {
			break
		}
		imageStr := ""
		for j := 0; j < 18; j++ {
			for i := 0; i < 32; i++ {
				pixel := buff[32*j+i]
				imageStr += asciiChar(pixel) + asciiChar(pixel)
			}
			imageStr += "\n"
		}
		currentMessage = header + imageStr
	}
}
